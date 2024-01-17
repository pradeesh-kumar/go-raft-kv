package kv

import (
	"encoding/json"
	"fmt"
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"sync"
	"time"
)

type KVStateMachine struct {
	inMemMap           map[string]string
	lastPersistedIndex uint64
	persistTimer       time.Timer
	saveMu             sync.Mutex
}

func newKVStateMachine() *KVStateMachine {
	return &KVStateMachine{
		lastPersistedIndex: 0,
	}
}

func (s *KVStateMachine) Apply(logs []*raft.StateMachineEntry) {
	for _, e := range logs {
		kvCommand, err := parseCommand(e.Value)
		if err != nil {
			logger.Error(err)
			continue
		}
		switch cmd := kvCommand.(type) {
		case *InsertCommand:
			s.processInsertCommand(cmd)
		case *DeleteCommand:
			s.processDeleteCommand(cmd)
		default:
			logger.Error("Unrecognized command ", cmd)
		}
	}
}

func (s *KVStateMachine) CaptureSnapshot(snapshotWriter raft.SnapshotWriter) error {
	serializedMap, err := json.Marshal(s.inMemMap)
	if err != nil {
		return fmt.Errorf("failed to take snapshot %s", err)
	}
	return snapshotWriter.Write(serializedMap)
}

func (s *KVStateMachine) ResetFromSnapshot(reader raft.SnapshotReader) error {
	// TODO implement this
	return nil
}

func (s *KVStateMachine) Get(key string) string {
	return s.inMemMap[key]
}

func (s *KVStateMachine) processInsertCommand(cmd *InsertCommand) {
	s.inMemMap[cmd.key] = cmd.val
}

func (s *KVStateMachine) processDeleteCommand(cmd *DeleteCommand) {
	delete(s.inMemMap, cmd.key)
}

func (s *KVStateMachine) persist() {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()
}

func (s *KVStateMachine) Close() {
	s.persistTimer.Stop()
	s.persist()
}
