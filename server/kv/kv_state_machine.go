package kv

import (
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
		insertCommand := NewInsertCommandFromBytes(e.Value)
		s.Put(insertCommand.key, insertCommand.val)
	}
}

func (s *KVStateMachine) Put(key string, val string) {
	s.inMemMap[key] = val
}

func (s *KVStateMachine) Get(key string) string {
	return s.inMemMap[key]
}

func (s *KVStateMachine) persist() {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()
}

func (s *KVStateMachine) Close() {
	s.persistTimer.Stop()
	s.persist()
}
