package main

import (
	"context"
	"flag"
	"github.com/pradeesh-kumar/go-raft-kv/kv"
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"github.com/pradeesh-kumar/go-raft-kv/rlog"
	"os"
	"os/signal"
	"time"
)

var (
	serverId      = flag.String("serverId", "node_1", "Server id")
	serverAddress = flag.String("serverAddress", "localhost:7071", "Address where the server listens to (localhost:7071)")
)

func main() {
	flag.Parse()
	nodes := make(map[raft.ServerId]raft.ServerAddress)
	nodes["node_1"] = "localhost:7071"
	nodes["node_2"] = "localhost:7072"
	nodes["node_3"] = "localhost:7073"

	kvConfig := &kv.KVConfig{
		RaftConfig: raft.RaftConfig{
			ServerId:             raft.ServerId(*serverId),
			ServerAddress:        raft.ServerAddress(*serverAddress),
			BaseDirectory:        "/tmp/goraft/",
			ReplicationBatchSize: 50,
			Timeouts: raft.Timeouts{
				ElectionTimeout:  time.Duration(4000) * time.Millisecond,
				BroadcastTimeout: time.Duration(1400) * time.Millisecond,
				HeartbeatTimeout: time.Duration(1500) * time.Millisecond,
				BatchDelay:       time.Duration(50) * time.Millisecond,
				ElectionJitter:   500,
			},
			Nodes: nodes,
		},
		LogConfig: rlog.LogConfig{
			MaxStoreBytes: 1048576,
			MaxIndexBytes: 1048576,
			InitialOffset: 1,
		},
	}
	kvServer := kv.NewKVServer(kvConfig)
	go func() {
		kvServer.Start()
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	kvServer.Stop()
}
