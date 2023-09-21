package raft

import (
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"math"
	"math/rand"
	"time"
)

type RaftServerImpl struct {
	config   *RaftConfig
	serverId ServerId
	state    RaftState

	// Persistent state
	currentTerm uint32   // Latest term server has seen
	votedFor    ServerId // CandidateId that received vote in current term (or null if none). This is nothing but the current leader

	// Volatile state
	currentLeader    ServerId // Server Id of the current Leader
	commitIndex      uint64   // Index of highest log entry known to be committed
	lastLogTerm      uint32   // Term of the highest log entry known to be committed
	lastAppliedIndex uint64   // Index of highest log applied to state machine

	// Used by leader to maintain the log indices of followers
	nextIndex  map[ServerId]uint64 // for each server, index of next log entry to send to that server
	matchIndex map[ServerId]uint64 // for each server, index of highest log entry known to be replicated on server

	leadershipTransferFlag bool

	raftLog      RaftLog
	transport    Transport
	stateMachine StateMachine
	stateStorage StateStorage

	electionTimer *time.Timer
	stopChannel   chan bool
	rpcChan       chan *RPC
	offerChan     chan *OfferRequest
}

func NewRaftServer(cfg RaftConfig, raftLog RaftLog, stateStorage StateStorage, sm StateMachine, t Transport) *RaftServerImpl {
	persistedState, err := stateStorage.Retrieve()
	if err != nil {
		logger.Fatal("Failed to retrieve persisted state: ", err)
	}
	raftServer := &RaftServerImpl{
		config:           &cfg,
		serverId:         cfg.ServerId,
		raftLog:          raftLog,
		stateStorage:     stateStorage,
		stateMachine:     sm,
		transport:        t,
		electionTimer:    time.NewTimer(largeFuture),
		currentTerm:      persistedState.CurrentTerm,
		votedFor:         ServerId(persistedState.VotedFor),
		lastAppliedIndex: persistedState.LastAppliedIndex,
		commitIndex:      persistedState.CommittedIndex,
		stopChannel:      make(chan bool),
		rpcChan:          make(chan *RPC),
		offerChan:        make(chan *OfferRequest, cfg.ReplicationBatchSize),
	}
	raftServer.changeState(newFollowerState(raftServer))
	raftServer.transport.RegisterMessageHandler(raftServer)
	highestOffset, err := raftServer.raftLog.HighestOffset()
	if err != nil {
		logger.Fatal(err)
	}
	raftServer.commitIndex = highestOffset
	logger.Debug("Commit index: ", raftServer.commitIndex)
	if raftServer.commitIndex > 0 {
		record, err := raftServer.raftLog.Read(raftServer.commitIndex)
		if err != nil {
			logger.Fatal(err)
		}
		raftServer.lastLogTerm = record.Term
	}
	return raftServer
}

func (r *RaftServerImpl) Start() {
	logger.Debug("Starting Raft Server")
	r.changeState(newFollowerState(r))
	r.transport.Start()
	r.scheduleElection()
	logger.Debug("Started")
	for {
		select {
		case <-r.electionTimer.C:
			r.startElection()
		case rpc := <-r.rpcChan:
			r.state.handleRPC(rpc)
		case cmd := <-r.offerChan:
			batch := []*OfferRequest{cmd}
		batchLoop:
			for i := 0; i < r.config.ReplicationBatchSize; i++ {
				select {
				case newLog := <-r.offerChan:
					batch = append(batch, newLog)
				default:
					break batchLoop
				}
			}
			r.state.handleRPC(&RPC{batch, nil})
		case <-r.stopChannel:
			logger.Info("Received stop signal. Terminating the Raft Server.")
			return
		}
	}
}

func (r *RaftServerImpl) Stop() {
	r.electionTimer.Stop()
	r.state.stop()
	r.stopChannel <- true
}

func (r *RaftServerImpl) HandleRPC(cmd any) Future {
	future := NewBlockingFuture()
	r.rpcChan <- &RPC{cmd: cmd, reply: future}
	return future
}

func (r *RaftServerImpl) OfferCommand(cmd Command) Future {
	future := NewBlockingFuture()
	r.offerChan <- &OfferRequest{cmd: cmd, reply: future}
	return future
}

func (r *RaftServerImpl) changeState(newState RaftState) {
	if r.state != nil {
		r.state.stop()
	}
	r.state = newState
	r.state.start()
}

// StartElection Begins leader election and sends the request to all the nodes in the cluster
func (r *RaftServerImpl) startElection() {
	r.currentLeader = ""
	r.changeState(newCandidateState(r))
	r.votedFor = r.serverId
	r.currentTerm++
	voteRequest := &VoteRequest{
		Term:         r.currentTerm,
		CandidateId:  string(r.serverId),
		LastLogIndex: r.commitIndex,
		LastLogTerm:  r.lastLogTerm,
	}
	logger.Infof("Election started for the new term %d", r.currentTerm)
	var voteBroadcastReq BroadcastRequest[*VoteRequest]
	for nodeId, nodeAddress := range r.config.Nodes {
		if nodeId == r.serverId {
			continue
		}
		voteBroadcastReq = append(voteBroadcastReq, NewPayload[*VoteRequest](nodeId, nodeAddress, voteRequest))
	}
	voteBroadcastResponse := r.transport.BroadcastVote(voteBroadcastReq)
	positiveVotes := 1 // Including self vote
	stale := false
	for _, response := range voteBroadcastResponse {
		if response.Message.VoteGranted {
			positiveVotes++
		} else if response.Message.Term > r.currentTerm {
			logger.Infof("Election failed since node %s has greater term %d than the current node term %d", response.ServerId, response.Message.Term, r.currentTerm)
			r.currentTerm = response.Message.Term
			stale = true
			break
		}
	}
	quorumMajority := r.getQuorumMajority()
	elected := !stale && positiveVotes >= quorumMajority

	if elected {
		logger.Infof("Server %s won the election in the term %d", r.serverId, r.currentTerm)
		r.changeState(newLeaderState(r))
		r.currentLeader = r.serverId
	} else {
		logger.Infof("Quorum majority %d positive votes %d", quorumMajority, positiveVotes)
		logger.Infof("Server %s lost the election", r.serverId)
		r.votedFor = ""
		r.changeState(newFollowerState(r))
		r.scheduleElection()
	}
	r.persistState()
}

func (r *RaftServerImpl) getQuorumMajority() int {
	return (len(r.config.Nodes) + 1) / 2
}

func (r *RaftServerImpl) processVoteRequest(req *VoteRequest) (*VoteResponse, error) {
	r.stopElection()
	defer r.scheduleElection()

	logger.Infof("Received vote request %+v", req)
	var voteGranted bool
	if req.Term > r.currentTerm {
		r.currentTerm = req.Term
		r.changeState(newFollowerState(r))
		r.votedFor = ""
	}
	currentTermCondition := req.Term == r.currentTerm
	votedForCondition := r.votedFor == "" || r.votedFor == ServerId(req.CandidateId)
	logCondition := req.LastLogTerm > r.lastLogTerm || (req.LastLogTerm == r.lastLogTerm && req.LastLogIndex >= r.commitIndex)

	voteGranted = currentTermCondition && votedForCondition && logCondition
	if voteGranted {
		r.votedFor = ServerId(req.CandidateId)
	}
	return &VoteResponse{Term: r.currentTerm, VoteGranted: voteGranted}, nil
}

func (r *RaftServerImpl) processAppendEntriesRequest(req *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	r.stopElection()
	defer r.scheduleElection()
	logger.Infof("Received append entries request %+v", req)
	if req.Term < r.currentTerm {
		return &AppendEntriesResponse{Term: r.currentTerm, Success: false}, nil
	} else if req.Term >= r.currentTerm || r.votedFor != "" {
		r.currentTerm = req.Term
		r.changeState(newFollowerState(r))
		r.votedFor = ""
		r.currentLeader = ServerId(req.LeaderId)
	}

	if len(req.Entries) > 0 && !r.matchEntry(req.PrevLogIndex, req.PrevLogTerm) {
		return &AppendEntriesResponse{Term: r.currentTerm, Success: false}, nil
	}

	newEntriesLen := len(req.Entries)
	if newEntriesLen > 0 && r.commitIndex > req.PrevLogIndex && req.PrevLogIndex > 0 {
		index := uint64(math.Min(float64(r.commitIndex), float64(req.PrevLogIndex+uint64(newEntriesLen))) - 1)
		record, _ := r.raftLog.Read(index)
		if record.Term != req.Entries[index-req.PrevLogIndex].Term {
			err := r.raftLog.Truncate(req.PrevLogIndex - 1)
			if err != nil {
				logger.Info("Error during truncate with index: ", req.PrevLogIndex-1)
			}
		}
	}
	if req.PrevLogIndex+uint64(len(req.Entries)) > r.commitIndex {
		for i := r.commitIndex - req.PrevLogIndex; i < uint64(len(req.Entries)); i++ {
			_, err := r.raftLog.Append(&Record{
				Term:  req.Entries[i].Term,
				Value: req.Entries[i].Value,
			})
			if err != nil {
				logger.Infof("Error while appending the entry %+v\n", req.Entries[i])
			}
			r.commitIndex++
		}
		logger.Infof("Committed %d entries to the state machine", len(req.Entries))
	}
	if newEntriesLen > 0 {
		r.persistState()
	}
	return &AppendEntriesResponse{Term: r.currentTerm, Success: true}, nil
}

func (r *RaftServerImpl) transferLeadership(*TransferLeadershipRequest) (*TransferLeadershipResponse, error) {
	response := &TransferLeadershipResponse{
		Success:  false,
		LeaderId: string(r.currentLeader),
	}
	if r.leadershipTransferFlag {
		return response, nil
	}
	r.leadershipTransferFlag = true
	defer func() {
		r.leadershipTransferFlag = false
	}()

	// Find the right candidate
	var candidate ServerId
	for serverId, index := range r.matchIndex {
		if index >= r.commitIndex {
			candidate = serverId
			break
		}
	}
	if candidate == "" {
		return response, nil
	}
	timeoutReq := &TimeoutNowRequest{
		LeaderId: string(r.serverId),
	}
	reqPayload := NewPayload[*TimeoutNowRequest](candidate, r.config.Nodes[candidate], timeoutReq)
	res, err := r.transport.SendTimeoutRequest(reqPayload)
	if err != nil {
		return nil, err
	}
	response.Success = res.Message.Success
	return response, nil
}

func (r *RaftServerImpl) timeoutNow(req *TimeoutNowRequest) (*TimeoutNowResponse, error) {
	response := &TimeoutNowResponse{Success: false}
	if ServerId(req.LeaderId) != r.currentLeader {
		return response, nil
	}
	r.startElection()
	response.Success = r.state.name() == State_Leader
	return response, nil
}

func (r *RaftServerImpl) matchEntry(prevLogIndex uint64, prevLogTerm uint32) bool {
	if r.commitIndex < prevLogIndex {
		return false
	}
	if prevLogIndex == 0 {
		return true
	}
	record, _ := r.raftLog.Read(prevLogIndex)
	return record.Term == prevLogTerm
}

func (r *RaftServerImpl) scheduleElection() {
	jitter := time.Millisecond * time.Duration(rand.Int63n(r.config.Timeouts.ElectionJitter))
	r.electionTimer.Reset(r.config.Timeouts.ElectionTimeout + jitter)
}

func (r *RaftServerImpl) stopElection() {
	r.electionTimer.Stop()
}

func (r *RaftServerImpl) persistState() {
	go func() {
		state := &PersistentState{
			CurrentTerm:      r.currentTerm,
			VotedFor:         string(r.votedFor),
			CommittedIndex:   r.commitIndex,
			LastAppliedIndex: r.lastAppliedIndex,
		}
		if err := r.stateStorage.Store(state); err != nil {
			logger.Errorf("Failed to store the state %+v", state)
		}
	}()
}
