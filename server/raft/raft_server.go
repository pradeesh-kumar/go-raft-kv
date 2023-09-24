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

	raftLog      RaftLog
	transport    Transport
	stateMachine StateMachine
	stateStorage StateStorage

	electionTimer *time.Timer
	stopChannel   chan bool
	applySignalCh chan any
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
	raftServer.commitIndex = 0
	if raftServer.raftLog.HighestOffset() > 0 {
		record, err := raftServer.raftLog.Read(raftServer.raftLog.HighestOffset())
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
		case <-r.applySignalCh:
			r.applyLogs()
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
			r.persistState()
			return
		}
	}
}

func (r *RaftServerImpl) Stop() {
	r.electionTimer.Stop()
	r.state.stop()
	r.stopChannel <- true
}

func (r *RaftServerImpl) State() State {
	return r.state.name()
}

func (r *RaftServerImpl) currentConfig() *ConfigEntry {
	var nodeInfos []*NodeInfo
	for serverId, serverAddress := range r.config.Nodes {
		nodeInfo := &NodeInfo{
			NodeId:  string(serverId),
			Address: string(serverAddress),
		}
		nodeInfos = append(nodeInfos, nodeInfo)
	}
	return &ConfigEntry{
		Nodes: nodeInfos,
	}
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
	if r.state.name() == State_Learner {
		r.scheduleElection()
		logger.Warn("Learner cannot start election")
		return
	}
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
		r.currentLeader = r.serverId
		r.changeState(newLeaderState(r))
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

func (r *RaftServerImpl) applyLogs() (logs []*Record, err error) {
	commitIndex := r.commitIndex
	lastAppliedIndex := r.lastAppliedIndex
	offset := lastAppliedIndex + 1
	batchSize := int(commitIndex - offset)
	if batchSize < 1 {
		logger.Warn("Nothing to apply")
		return
	}
	logs, err = r.raftLog.ReadBatchSince(offset, batchSize)
	if err != nil {
		logger.Errorf("Failed to retrieve the logs for apply. From index: %d, batchSize: %d", offset, batchSize, err)
		return
	}
	var stateMachineEntries []*StateMachineEntry
	var configEntry *ConfigEntry = nil
	for i, log := range logs {
		switch entry := log.LogEntryBody.(type) {
		case *Record_ConfigEntry:
			configEntry = entry.ConfigEntry
		case *Record_StateMachineEntry:
			stateMachineEntries = append(stateMachineEntries, entry.StateMachineEntry)
		case *Record_SnapshotMetadataEntry:
			r.stateMachine.Apply(stateMachineEntries)
			stateMachineEntries = make([]*StateMachineEntry, 0)
			err := r.stateMachine.TakeSnapshot()
			if err != nil {
				logger.Errorf("Failed to take snapshot ", err)
				r.lastAppliedIndex = log.Offset - 1
				r.persistState()
				if configEntry != nil {
					r.updateLatestConfig(configEntry)
				}
				return logs[:i-1], nil
			} else {
				err := r.raftLog.Truncate(log.Offset - 1)
				if err != nil {
					logger.Errorf("failed to truncate logs ", err)
				}
			}
		}
	}
	r.stateMachine.Apply(stateMachineEntries)
	if configEntry != nil {
		r.updateLatestConfig(configEntry)
	}
	r.lastAppliedIndex = commitIndex
	r.persistState()
	return
}

func (r *RaftServerImpl) updateLatestConfig(configEntry *ConfigEntry) {
	r.config.Nodes = make(map[ServerId]ServerAddress)
	for _, nodeInfo := range configEntry.Nodes {
		serverId := ServerId(nodeInfo.NodeId)
		serverAddress := ServerAddress(nodeInfo.Address)
		r.config.Nodes[serverId] = serverAddress
	}
	if _, ok := r.config.Nodes[r.serverId]; !ok && r.state.name() == State_Learner {
		r.changeState(newFollowerState(r))
	}
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
		return &AppendEntriesResponse{Term: r.currentTerm, LastLogIndex: r.raftLog.HighestOffset(), Success: false}, nil
	} else if req.Term >= r.currentTerm || r.votedFor != "" {
		r.currentTerm = req.Term
		r.changeState(newFollowerState(r))
		r.votedFor = ""
		r.currentLeader = ServerId(req.LeaderId)
	}

	if len(req.Entries) > 0 && !r.matchEntry(req.PrevLogIndex, req.PrevLogTerm) {
		return &AppendEntriesResponse{Term: r.currentTerm, LastLogIndex: r.raftLog.HighestOffset(), Success: false}, nil
	}
	newEntriesLen := len(req.Entries)
	currentLogIndex := r.raftLog.HighestOffset()
	if newEntriesLen > 0 && currentLogIndex > req.PrevLogIndex && req.PrevLogIndex > 0 {
		index := uint64(math.Min(float64(currentLogIndex), float64(req.PrevLogIndex+uint64(newEntriesLen))) - 1)
		record, _ := r.raftLog.Read(index)
		if record.Term != req.Entries[index-req.PrevLogIndex].Term {
			err := r.raftLog.Truncate(req.PrevLogIndex - 1)
			if err != nil {
				logger.Info("Error during truncate with index: ", req.PrevLogIndex-1)
			}
		}
	}
	if req.PrevLogIndex+uint64(len(req.Entries)) > currentLogIndex {
		for i := r.commitIndex - req.PrevLogIndex; i < uint64(len(req.Entries)); i++ {
			_, err := r.raftLog.Append(req.Entries[i])
			if err != nil {
				logger.Infof("Error while appending the entry %+v\n", req.Entries[i])
			}
		}
		logger.Infof("Committed %d entries to the state machine", len(req.Entries))
	}
	if req.LeaderCommitIndex > r.commitIndex {
		r.commitIndex = req.LeaderCommitIndex
	}
	if newEntriesLen > 0 || req.LeaderCommitIndex > r.commitIndex {
		r.persistState()
	}
	return &AppendEntriesResponse{Term: r.currentTerm, LastLogIndex: r.raftLog.HighestOffset(), Success: true}, nil
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
