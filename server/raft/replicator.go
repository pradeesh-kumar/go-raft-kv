package raft

import (
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"time"
)

type Replicator interface {
	signal()
	start()
	stop()
	followerInfo() *Follower
}

type DefaultReplicator struct {
	follower                *Follower
	leaderState             *LeaderState
	raftServer              *RaftServerImpl
	transport               Transport
	heartBeatTimeout        time.Duration
	batchDelay              time.Duration
	heartBeatTimer          *time.Timer
	followerAppendSuccessCh chan<- any
	batchSize               int
	signalCh                chan any
	stopCh                  chan any
	busy                    bool
	running                 bool
}

func newDefaultReplicator(follower *Follower, leaderState *LeaderState, followerAppendSuccessCh chan<- any) *DefaultReplicator {
	defaultReplicator := &DefaultReplicator{
		follower:                follower,
		leaderState:             leaderState,
		raftServer:              leaderState.raftServer,
		transport:               leaderState.raftServer.transport,
		batchSize:               leaderState.raftServer.config.ReplicationBatchSize,
		heartBeatTimeout:        leaderState.raftServer.config.Timeouts.HeartbeatTimeout,
		batchDelay:              leaderState.raftServer.config.Timeouts.BatchDelay,
		heartBeatTimer:          time.NewTimer(largeFuture),
		followerAppendSuccessCh: followerAppendSuccessCh,
		signalCh:                make(chan any),
		stopCh:                  make(chan any),
	}
	return defaultReplicator
}

func (r *DefaultReplicator) signal() {
	r.signalCh <- true
}

func (r *DefaultReplicator) start() {
	r.startInBackground()
}

func (r *DefaultReplicator) startInBackground() {
	go func() {
		for {
			select {
			case <-r.signalCh:
				r.sendHeartbeat()
			case <-r.heartBeatTimer.C:
				r.sendHeartbeat()
			case <-r.stopCh:
				logger.Debug("Replicator received stop signal. Terminating the replicator")
				r.stopHeartbeat()
				r.running = false
				return
			}
		}
	}()
	r.sendHeartbeat()
}

func (r *DefaultReplicator) stop() {
	r.stopCh <- true
}

func (r *DefaultReplicator) stopHeartbeat() {
	r.heartBeatTimer.Stop()
}

func (r *DefaultReplicator) scheduleHeartbeat() {
	r.heartBeatTimer.Reset(r.heartBeatTimeout)
}

func (r *DefaultReplicator) scheduleImmediateNext() {
	r.heartBeatTimer.Reset(r.batchDelay)
}

func (r *DefaultReplicator) followerInfo() *Follower {
	return r.follower
}

func (r *DefaultReplicator) sendHeartbeat() {
	if r.busy {
		return
	}
	r.busy = true
	defer func() {
		r.busy = false
	}()
	logger.Info("Heartbeat triggered for the node")
	prevLogIndex := r.follower.nextIndex - 1
	var prevLogTerm uint32
	if prevLogIndex > 0 {
		record, _ := r.raftServer.raftLog.Read(prevLogIndex)
		prevLogTerm = record.Term
	}
	var records []*Record
	records, err := r.raftServer.raftLog.ReadBatchSince(r.follower.nextIndex, r.batchSize)
	if err != nil {
		logger.Errorf("Error while retrieving records to send append-entries", err)
		r.scheduleHeartbeat()
		return
	}
	appendEntryReq := &AppendEntriesRequest{
		Term:              r.raftServer.currentTerm,
		LeaderId:          string(r.raftServer.currentLeader),
		PrevLogIndex:      prevLogIndex,
		PrevLogTerm:       prevLogTerm,
		LeaderCommitIndex: r.raftServer.commitIndex,
		Entries:           records,
	}
	reqPayload := NewPayload[*AppendEntriesRequest](r.follower.id, r.follower.address, appendEntryReq)
	response, err := r.transport.SendAppendEntries(reqPayload)
	if err != nil {
		logger.Errorf("Failed to send append entries to the node %d", r.follower.id, err)
		return
	}
	if !r.running {
		return
	}
	recordsLength := len(records)
	if response.Message.Success {
		r.follower.nextIndex = r.follower.nextIndex + uint64(recordsLength)
		r.follower.matchIndex = r.follower.nextIndex - 1
		logger.Infof("Appended the %d entries to the node %d", recordsLength, r.follower.id)
		r.followerAppendSuccessCh <- true
	}
	if response.Message.Term > r.raftServer.currentTerm {
		logger.Infof("Leader stepping down since node %s has greater term %d than the current node term %d", response.ServerId, response.Message.Term, r.raftServer.currentTerm)
		r.raftServer.currentTerm = response.Message.Term
		r.leaderState.stepDown()
		return
	} else {
		r.follower.nextIndex = r.follower.nextIndex - 1
	}
	logger.Infof("Node %s is behind the commit", response.ServerId)
	if recordsLength > 1 {
		r.scheduleImmediateNext()
	} else {
		r.scheduleHeartbeat()
	}
}
