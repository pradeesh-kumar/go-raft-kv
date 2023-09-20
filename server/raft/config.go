package raft

import (
	"time"
)

type ServerId string
type ServerAddress string

type RaftConfig struct {
	ServerId             ServerId                   `yaml:"serverId"`
	ServerAddress        ServerAddress              `yaml:"raftListenAddress"`
	Timeouts             Timeouts                   `yaml:"timeouts"`
	BaseDirectory        string                     `yaml:"baseDirectory"`
	ReplicationBatchSize int                        `yaml:"replicationBatchSize"`
	Nodes                map[ServerId]ServerAddress `yaml:"nodes"`
}

type Timeouts struct {
	ElectionTimeout  time.Duration `yaml:"electionTimeout"`
	HeartbeatTimeout time.Duration `yaml:"heartbeatTimeout"`
	BatchDelay       time.Duration `yaml:"batchDelay"`
	BroadcastTimeout time.Duration `yaml:"broadcastTimeout"`
	ElectionJitter   int64         `yaml:"electionJitter"`
}
