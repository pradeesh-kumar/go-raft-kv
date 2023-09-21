package raft

import (
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"math"
)

const DefaultChanBuffer = 1000

type LeaderState struct {
	raftServer              *RaftServerImpl
	replicators             map[ServerId]Replicator
	pendingRequests         map[uint64]MutableFuture
	followerAppendSuccessCh <-chan any
	applySignalCh           chan any
	stopChannel             chan any
}

func newLeaderState(r *RaftServerImpl) RaftState {
	leaderState := &LeaderState{
		raftServer:      r,
		pendingRequests: make(map[uint64]MutableFuture),
		stopChannel:     make(chan any),
		applySignalCh:   make(chan any, DefaultChanBuffer),
	}
	followerAppendSuccessCh := make(chan any, DefaultChanBuffer)
	replicators := make(map[ServerId]Replicator)
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
	// TODO create an empty config log amd Replicate it
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
			for _, r := range s.replicators {
				r.stop()
			}
			for _, pendingRequest := range s.pendingRequests {
				pendingRequest.Set(&OfferResponse{Status: ResponseStatus_LeaderStepDown}, nil)
			}
			return
		}
	}
}

func (s *LeaderState) updateCommitIndex() {
	var minAppendIndex uint64 = math.MaxUint64
	for _, repl := range s.replicators {
		minAppendIndex = uint64(math.Min(float64(minAppendIndex), float64(repl.followerInfo().matchIndex)))
	}
	if minAppendIndex > s.raftServer.commitIndex {
		s.raftServer.commitIndex = minAppendIndex
		s.raftServer.persistState()
		s.applySignalCh <- true
	}
}

func (s *LeaderState) applyLogs() {
	commitIndex := s.raftServer.commitIndex
	lastAppliedIndex := s.raftServer.lastAppliedIndex
	offset := lastAppliedIndex + 1
	batchSize := int(commitIndex - offset)
	if batchSize < 1 {
		logger.Warn("Nothing to apply")
		return
	}
	logs, err := s.raftServer.raftLog.ReadBatchSince(offset, batchSize)
	if err != nil {
		logger.Errorf("Failed to retrieve the logs for apply. From index: %d, batchSize: %d", offset, batchSize, err)
		return
	}
	s.raftServer.stateMachine.Apply(logs)
	s.raftServer.lastAppliedIndex = commitIndex
	s.raftServer.persistState()
	s.replyPendingRequests(logs)
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

func (s *LeaderState) handleRPC(rpc *RPC) {
	switch cmd := rpc.cmd.(type) {
	case *VoteRequest:
		rpc.reply.Set(s.raftServer.processVoteRequest(cmd))
	case *AppendEntriesRequest:
		rpc.reply.Set(s.raftServer.processAppendEntriesRequest(cmd))
	case *TransferLeadershipRequest:
		rpc.reply.Set(s.raftServer.transferLeadership(cmd))
	case *TimeoutNowRequest:
		rpc.reply.Set(&TimeoutNowResponse{Success: false}, nil)
	case *AddServerRequest:
		// TODO implement
	case *RemoveServerRequest:
		// TODO implement
	case []*OfferRequest:
		s.offerCommand(cmd)
	default:
		rpc.reply.Set(nil, ErrUnknownCommand)
	}
}

// Appends the logs to the local log, replicates, and commits
func (s *LeaderState) offerCommand(requests []*OfferRequest) {
	// TODO - Reject requests during leadership transfer
	for _, offerReq := range requests {
		r := &Record{
			Term:  s.raftServer.currentTerm,
			Value: offerReq.cmd.Bytes(),
		}
		id, err := s.raftServer.raftLog.Append(r)
		if err != nil {
			logger.Errorf("Failed to append the new record %+v", r, err)
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
	s.raftServer.changeState(newFollowerState(s.raftServer))
}

func (s *LeaderState) triggerReplicators() {
	for _, r := range s.replicators {
		r.signal()
	}
}

func (s *LeaderState) stop() {
	s.stopChannel <- true
}
