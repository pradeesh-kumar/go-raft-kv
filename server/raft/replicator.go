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
				return
			}
		}
	}()
	r.sendHeartbeat()
}

func (r *DefaultReplicator) stop() {
	logger.Debug("Replicator received stop signal. Terminating the replicator")
	r.stopHeartbeat()
	r.running = false
	r.stopCh <- true
}

func (r *DefaultReplicator) stopHeartbeat() {
	r.heartBeatTimer.Stop()
}

func (r *DefaultReplicator) scheduleHeartbeat() {
	r.heartBeatTimer.Reset(r.heartBeatTimeout)
}

func (r *DefaultReplicator) scheduleImmediateNext() {
	r.heartBeatTimer.Reset(0)
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
		logger.Error("Error while retrieving records to send append-entries ", err)
		r.scheduleHeartbeat()
		return
	}
	if containsSnapshotEntry(records) {
		r.sendSnapshot(records[0].LogEntryBody.(*Record_SnapshotMetadataEntry).SnapshotMetadataEntry)
	} else {
		r.sendLogs(records, prevLogIndex, prevLogTerm)
	}
}

func (r *DefaultReplicator) sendLogs(records []*Record, prevLogIndex uint64, prevLogTerm uint32) {
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
		logger.Errorf("Failed to send append entries to the node %s: %s", r.follower.id, err)
		return
	}
	if !r.running {
		return
	}
	recordsLength := len(records)
	if response.Message.Term > r.raftServer.currentTerm {
		logger.Infof("Leader stepping down since node %s has greater term %d than the current node term %d", response.ServerId, response.Message.Term, r.raftServer.currentTerm)
		r.raftServer.currentTerm = response.Message.Term
		r.leaderState.stepDown()
		return
	}
	if response.Message.Success {
		r.follower.nextIndex = r.follower.nextIndex + uint64(recordsLength)
		r.follower.matchIndex = r.follower.nextIndex - 1
		logger.Infof("Appended the %d entries to the node %s", recordsLength, r.follower.id)
		r.followerAppendSuccessCh <- true
	} else {
		if response.Message.LastLogIndex < r.follower.nextIndex {
			r.follower.nextIndex = response.Message.LastLogIndex
		} else {
			r.follower.nextIndex = r.follower.nextIndex - 1
		}
		logger.Infof("Node %s is behind the commit", response.ServerId)
	}
	if recordsLength > 1 {
		r.scheduleImmediateNext()
	} else {
		r.scheduleHeartbeat()
	}
}

func (r *DefaultReplicator) sendSnapshot(snapshotMetadata *SnapshotMetadataEntry) {
	snapshot, err := r.raftServer.snapshotManager.GetLatestSnapshot()
	if err != nil {
		logger.Error("Failed to retrieve the latest snapshot: ", err)
	}
	bytes, err := snapshot.ReadAll()
	if err != nil {
		logger.Error("Failed to read the latest snapshot: ", err)
	}
	installSnapshotReq := &InstallSnapshotRequest{
		Term:         r.raftServer.currentTerm,
		LeaderId:     string(r.raftServer.currentLeader),
		LastLogIndex: snapshotMetadata.CommitIndex,
		LastLogTerm:  snapshotMetadata.CommitLogTerm,
		LastConfig:   snapshotMetadata.LastConfigIndex,
		Offset:       0,
		Value:        bytes,
		Done:         true,
	}
	reqPayload := NewPayload[*InstallSnapshotRequest](r.follower.id, r.follower.address, installSnapshotReq)
	response, err := r.transport.SendInstallSnapshot(reqPayload)
	if err != nil {
		logger.Errorf("Failed to send snapshot entry to the node %s: %s", r.follower.id, err)
		return
	}
	if !r.running {
		return
	}
	if response.Message.Term > r.raftServer.currentTerm {
		logger.Infof("Leader stepping down since node %s has greater term %d than the current node term %d", response.ServerId, response.Message.Term, r.raftServer.currentTerm)
		r.raftServer.currentTerm = response.Message.Term
		r.leaderState.stepDown()
		return
	}
	r.follower.nextIndex = snapshotMetadata.CommitIndex + 1
	r.follower.matchIndex = snapshotMetadata.CommitIndex
	logger.Infof("Successfully sent the snapshot %+v to the node %s", snapshotMetadata, r.follower.id)
	r.scheduleHeartbeat()
}

func containsSnapshotEntry(records []*Record) bool {
	if len(records) > 0 {
		_, ok := records[0].LogEntryBody.(*Record_SnapshotMetadataEntry)
		return ok
	}
	return false
}
