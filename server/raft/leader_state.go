package raft

import (
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"sort"
	"time"
)

const (
	defaultChanBuffer                   = 1000
	membershipMaxRounds                 = 5
	leadershipTransferLogCatchupTimeout = time.Duration(4000) * time.Millisecond
	logsPerSnapshot                     = 1000
)

type LeaderState struct {
	raftServer              *RaftServerImpl
	replicators             map[ServerId]Replicator
	pendingRequests         map[uint64]MutableFuture
	leadershipTransferFlag  bool
	running                 bool
	followerAppendSuccessCh chan any
	applySignalCh           chan any
	stopChannel             chan any
}

func newLeaderState(r *RaftServerImpl) RaftState {
	leaderState := &LeaderState{
		raftServer:      r,
		pendingRequests: make(map[uint64]MutableFuture),
		stopChannel:     make(chan any),
		applySignalCh:   make(chan any, defaultChanBuffer),
	}
	followerAppendSuccessCh := make(chan any, defaultChanBuffer)
	replicators := make(map[ServerId]Replicator)
	r.raftLog.HighestOffset()
	for serverId, serverAddress := range r.config.Nodes {
		node := Node{serverId, serverAddress}
		follower := &Follower{Node: node, nextIndex: r.commitIndex + 1, matchIndex: 0}
		replicators[serverId] = newDefaultReplicator(follower, leaderState, followerAppendSuccessCh)
	}
	leaderState.replicators = replicators
	leaderState.followerAppendSuccessCh = followerAppendSuccessCh
	return leaderState
}

func (*LeaderState) name() State {
	return State_Leader
}

func (s *LeaderState) start() {
	s.running = true
	configEntry := s.raftServer.currentConfig()
	_, err := s.appendConfigEntry(configEntry)
	if err != nil {
		logger.Fatal("failed to append the initial leader config entry", err)
	}

	for {
		select {
		case <-s.followerAppendSuccessCh:
		batchLoop:
			for i := 0; i < 100; i++ {
				select {
				case <-s.followerAppendSuccessCh:
				default:
					break batchLoop
				}
			}
			s.updateCommitIndex()
		case <-s.applySignalCh:
			s.applyLogs()
		case <-s.stopChannel:
			return
		}
	}
}

func (s *LeaderState) handleRPC(rpc *RPC) {
	switch cmd := rpc.cmd.(type) {
	case *VoteRequest:
		rpc.reply.Set(s.raftServer.processVoteRequest(cmd))
	case *AppendEntriesRequest:
		rpc.reply.Set(s.raftServer.processAppendEntriesRequest(cmd))
	case *TransferLeadershipRequest:
		go rpc.reply.Set(s.transferLeadership())
	case *TimeoutNowRequest:
		rpc.reply.Set(&TimeoutNowResponse{Success: false}, nil)
	case *AddServerRequest:
		go rpc.reply.Set(s.addServer(cmd))
	case *RemoveServerRequest:
		go rpc.reply.Set(s.removeServer(cmd))
	case []*OfferRequest:
		s.offerCommand(cmd)
	case *InstallSnapshotRequest:
		rpc.reply.Set(s.handleInstallSnapshotRequest(cmd))
	default:
		rpc.reply.Set(nil, ErrUnknownCommand)
	}
}

func (s *LeaderState) handleInstallSnapshotRequest(req *InstallSnapshotRequest) (*InstallSnapshotResponse, error) {
	// Step out if the leader is stale
	if req.Term > s.raftServer.currentTerm {
		logger.Infof("Leader stepping down since node %s has greater term %d than the current node term %d", req.LeaderId, req.Term, s.raftServer.currentTerm)
		s.stepDown()
	}
	return &InstallSnapshotResponse{Term: s.raftServer.currentTerm}, nil
}

func (s *LeaderState) appendConfigEntry(configEntry *ConfigEntry) (uint64, error) {
	record := &Record{
		Term: s.raftServer.currentTerm,
		LogEntryBody: &Record_ConfigEntry{
			ConfigEntry: configEntry,
		},
	}
	return s.raftServer.raftLog.Append(record)
}

func (s *LeaderState) removeServer(req *RemoveServerRequest) (*RemoveServerResponse, error) {
	serverIdToBeRemoved := ServerId(req.ServerId)
	if _, ok := s.raftServer.config.Nodes[serverIdToBeRemoved]; !ok {
		logger.Infof("Invalid RemoveServerRequest! The requested server %s not found in the configuration", req.ServerId)
		return &RemoveServerResponse{Status: ResponseStatus_NotFound, LeaderId: string(s.raftServer.currentLeader)}, nil
	}
	configEntry := s.raftServer.currentConfig()
	cfgIdx := 0
	for i, node := range configEntry.Nodes {
		if node.NodeId == req.ServerId {
			cfgIdx = i
			break
		}
	}
	configEntry.Nodes = append(configEntry.Nodes[:cfgIdx], configEntry.Nodes[cfgIdx+1:]...)
	index, err := s.appendConfigEntry(configEntry)
	if err != nil {
		return nil, err
	}
	logger.Infof("Config entry successfully created on the index %d", index)
	future := NewBlockingFuture()
	s.pendingRequests[index] = future
	s.triggerReplicators()
	_, err = future.Get()
	if err != nil {
		return nil, err
	}
	s.replicators[serverIdToBeRemoved].stop()
	delete(s.replicators, serverIdToBeRemoved)
	return &RemoveServerResponse{
		Status:   ResponseStatus_Success,
		LeaderId: string(s.raftServer.serverId),
	}, nil
}

func (s *LeaderState) addServer(req *AddServerRequest) (*AddServerResponse, error) {
	if s.canAddServer(ServerId(req.ServerId)) {
		logger.Warn("Either the server %s already exist or other server add in progress.", req.ServerId)
		return &AddServerResponse{Status: ResponseStatus_Conflict, LeaderId: string(s.raftServer.currentLeader)}, nil
	}
	node := Node{ServerId(req.ServerId), ServerAddress(req.ServerAddress)}
	follower := &Follower{Node: node, nextIndex: s.raftServer.commitIndex + 1, matchIndex: 0}
	s.replicators[follower.id] = newDefaultReplicator(follower, s, s.followerAppendSuccessCh)
	s.replicators[follower.id].start()
	success := false
	for round := 1; round <= membershipMaxRounds; round++ {
		logger.Infof("Starting the round %d for the learner node %s", round, follower.id)
		currentLogIndex := s.raftServer.raftLog.HighestOffset()
		time.Sleep(s.raftServer.config.Timeouts.ElectionTimeout)

		if s.replicators[follower.id].followerInfo().matchIndex >= currentLogIndex {
			logger.Infof("Learner node %s picked up logs within election timeout at round %d", follower.id, round)
			success = true
		}
	}
	if !success {
		// The target server couldn't catch up the log in max rounds on time
		logger.Infof("Learner node %s failed to pick up logs within election timeout at max rounds %d", follower.id, membershipMaxRounds)
		s.replicators[follower.id].stop()
		delete(s.replicators, follower.id)
		return &AddServerResponse{Status: ResponseStatus_Timeout, LeaderId: string(s.raftServer.currentLeader)}, nil
	}
	newNodeInfo := &NodeInfo{
		NodeId:  req.ServerId,
		Address: req.ServerAddress,
	}
	configEntry := s.raftServer.currentConfig()
	configEntry.Nodes = append(configEntry.Nodes, newNodeInfo)
	index, err := s.appendConfigEntry(configEntry)
	if err != nil {
		return nil, err
	}
	logger.Infof("Config entry successfully created on the index %d", index)
	future := NewBlockingFuture()
	s.pendingRequests[index] = future
	s.triggerReplicators()
	_, err = future.Get()
	if err != nil {
		return nil, err
	}
	return &AddServerResponse{
		Status:   ResponseStatus_Success,
		LeaderId: string(s.raftServer.serverId),
	}, nil
}

// canAddServer checks if the specified serverId already present in the config or another server add operation in progress
func (s *LeaderState) canAddServer(serverId ServerId) bool {
	if _, ok := s.raftServer.config.Nodes[serverId]; ok {
		return false
	}
	if _, ok := s.replicators[serverId]; ok {
		return false
	}
	return true
}

func (s *LeaderState) transferLeadership() (*TransferLeadershipResponse, error) {
	response := &TransferLeadershipResponse{
		Success:  false,
		LeaderId: string(s.raftServer.currentLeader),
	}
	if s.leadershipTransferFlag {
		return response, nil
	}
	// Find the right candidate
	var candidate ServerId
	var highestIndex uint64
	for _, replicator := range s.replicators {
		if highestIndex < replicator.followerInfo().matchIndex {
			candidate = replicator.followerInfo().id
			highestIndex = replicator.followerInfo().matchIndex
		}
	}
	if candidate == "" {
		return response, nil
	}
	s.leadershipTransferFlag = true
	if s.replicators[candidate].followerInfo().matchIndex < s.raftServer.raftLog.HighestOffset() {
		// Wait for the candidate to catch up the logs
		time.Sleep(leadershipTransferLogCatchupTimeout)
		if s.replicators[candidate].followerInfo().matchIndex < s.raftServer.raftLog.HighestOffset() {
			// The candidate couldn't catch up the logs on the expected duration, cancelling the leadership transfer
			s.leadershipTransferFlag = false
			return response, nil
		}
	}
	timeoutReq := &TimeoutNowRequest{
		LeaderId: string(s.raftServer.serverId),
	}
	reqPayload := NewPayload[*TimeoutNowRequest](candidate, s.raftServer.config.Nodes[candidate], timeoutReq)
	res, err := s.raftServer.transport.SendTimeoutRequest(reqPayload)
	if err != nil {
		return nil, err
	}
	response.Success = res.Message.Success
	if response.Success {
		s.stepDown()
	}
	return response, nil
}

func (s *LeaderState) updateCommitIndex() {
	var followersMatchIndex []uint64
	for _, repl := range s.replicators {
		followersMatchIndex = append(followersMatchIndex, repl.followerInfo().matchIndex)
	}
	sort.Slice(followersMatchIndex, func(i, j int) bool { return followersMatchIndex[i] > followersMatchIndex[j] })
	quorumMajority := s.raftServer.getQuorumMajority()
	majorityCommitIndex := followersMatchIndex[quorumMajority-1]
	if majorityCommitIndex > s.raftServer.commitIndex {
		s.raftServer.commitIndex = majorityCommitIndex
		if majorityCommitIndex%logsPerSnapshot == 0 {
			s.appendSnapshotEntry(majorityCommitIndex)
		}
		s.raftServer.persistState()
		s.applySignalCh <- true
	}
}

func (s *LeaderState) appendSnapshotEntry(commitIndex uint64) {
	recordAtCommitIndex, err := s.raftServer.raftLog.Read(commitIndex)
	if err != nil {
		logger.Errorf("failed to read the log at commit index %d", commitIndex)
	}
	record := &Record{
		Term: s.raftServer.currentTerm,
		LogEntryBody: &Record_SnapshotMetadataEntry{
			SnapshotMetadataEntry: &SnapshotMetadataEntry{
				CommitIndex:     commitIndex,
				CommitLogTerm:   recordAtCommitIndex.Term,
				LastConfigIndex: 0, // TODO add the last config index
			},
		},
	}
	index, err := s.raftServer.raftLog.Append(record)
	if err != nil {
		logger.Error("failed to create snapshot entry ", err)
	} else {
		logger.Infof("Snapshot entry created at index %d", index)
	}
}

func (s *LeaderState) applyLogs() {
	logs, err := s.raftServer.applyLogs()
	if err != nil {
		s.replyPendingRequests(logs)
	}
}

func (s *LeaderState) replyPendingRequests(logs []*Record) {
	pendingRequestCount := 0
	leaderId := s.raftServer.currentLeader
	defaultOfferResponse := &OfferResponse{
		Status:   ResponseStatus_Success,
		LeaderId: string(leaderId),
	}
	for _, log := range logs {
		if _, ok := s.pendingRequests[log.Offset]; !ok {
			s.pendingRequests[log.Offset].Set(defaultOfferResponse, nil)
			pendingRequestCount++
		}
	}
	logger.Infof("Replied to %d pending requests", pendingRequestCount)
}

func (s *LeaderState) offerCommand(requests []*OfferRequest) {
	for _, offerReq := range requests {
		if s.leadershipTransferFlag {
			offerReq.reply.Set(&OfferResponse{Status: ResponseStatus_LeadershipTransferInProgress}, nil)
			continue
		}
		if !s.running {
			offerReq.reply.Set(&OfferResponse{Status: ResponseStatus_ShutdownInProgress}, nil)
			continue
		}
		r := &Record{
			Term: s.raftServer.currentTerm,
			LogEntryBody: &Record_StateMachineEntry{
				StateMachineEntry: &StateMachineEntry{Value: offerReq.cmd.Bytes()},
			},
		}
		id, err := s.raftServer.raftLog.Append(r)
		if err != nil {
			logger.Errorf("failed to append the new record %+v: %s", r, err)
			offerReq.reply.Set(nil, err)
		} else {
			s.pendingRequests[id] = offerReq.reply
			logger.Infof("New Record appended at index %d", id)
		}
	}
	s.triggerReplicators()
}

func (s *LeaderState) stepDown() {
	s.raftServer.votedFor = ""
	s.raftServer.currentLeader = ""
	s.raftServer.persistState()
	for _, pendingRequest := range s.pendingRequests {
		pendingRequest.Set(&OfferResponse{Status: ResponseStatus_LeaderStepDown}, nil)
	}
	s.raftServer.scheduleElection()
	s.raftServer.changeState(newFollowerState(s.raftServer))
}

func (s *LeaderState) triggerReplicators() {
	for _, r := range s.replicators {
		r.signal()
	}
}

func (s *LeaderState) stop() {
	for _, r := range s.replicators {
		r.stop()
	}
	for _, pendingRequest := range s.pendingRequests {
		pendingRequest.Set(&OfferResponse{Status: ResponseStatus_LeaderStepDown}, nil)
	}
	s.running = false
	s.stopChannel <- true
}
