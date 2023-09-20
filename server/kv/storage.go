package kv

import (
	"github.com/golang/protobuf/proto"
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"io"
	"os"
	"sync"
)

type LocalFileStateStorage struct {
	stateFile *os.File
	mu        sync.Mutex
}

func newLocalFileStateStorage(stateFile string) (*LocalFileStateStorage, error) {
	f, err := os.OpenFile(stateFile, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	return &LocalFileStateStorage{
		stateFile: f,
	}, nil
}

func (s *LocalFileStateStorage) Store(state *raft.PersistentState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bytes, err := proto.Marshal(state)
	if err != nil {
		return err
	}
	if _, err = s.stateFile.Write(bytes); err != nil {
		return err
	}
	return nil
}

func (s *LocalFileStateStorage) Retrieve() (*raft.PersistentState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bytes, err := io.ReadAll(s.stateFile)
	if err != nil {
		return nil, err
	}
	state := &raft.PersistentState{}
	if proto.Unmarshal(bytes, state) != nil {
		return nil, err
	}
	return state, nil
}

func (s *LocalFileStateStorage) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stateFile.Close()
}

type SnapshotStorage struct {
}
