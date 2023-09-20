package kv

import (
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"github.com/pradeesh-kumar/go-raft-kv/rlog"
)

type KVConfig struct {
	raft.RaftConfig
	rlog.LogConfig
}
