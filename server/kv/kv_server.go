package kv

import (
	"context"
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"github.com/pradeesh-kumar/go-raft-kv/rlog"
	"os"
	"path"
)

type KVServer struct {
	raftServer   raft.RaftServer
	raftLog      *rlog.Log
	stateMachine *KVStateMachine
	stateStorage *LocalFileStateStorage
	UnimplementedClientProtocolServiceServer
}

func NewKVServer(kvConfig *KVConfig) *KVServer {
	kvServer := &KVServer{}
	transport := NewGrpcTransport(TransportOpts{
		address:          kvConfig.RaftConfig.ServerAddress,
		broadcastTimeout: kvConfig.Timeouts.BroadcastTimeout,
	})
	baseDir := path.Join(kvConfig.BaseDirectory, string(kvConfig.RaftConfig.ServerId))
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		logger.Fatal("Error while creating base directory", err)
	}
	transport.RegisterClientService(&ClientProtocolService_ServiceDesc, kvServer)
	kvServer.stateMachine = newKVStateMachine()
	raftLog, err := rlog.NewLog(path.Join(baseDir, "logs"), kvConfig.LogConfig)
	if err != nil {
		logger.Fatal("Failed to initialize raft log: ", err)
	}
	kvServer.raftLog = raftLog
	stateStorage, err := newLocalFileStateStorage(path.Join(baseDir, "state.dat"))
	if err != nil {
		logger.Fatal("Failed to initialize KVServer ", err)
	}
	kvServer.stateStorage = stateStorage
	raftServer := raft.NewRaftServer(kvConfig.RaftConfig, raftLog, stateStorage, kvServer.stateMachine, transport)
	kvServer.raftServer = raftServer
	return kvServer
}

func (ks *KVServer) Start() {
	logger.Info("Starting Key-Value Server")
	ks.raftServer.Start()
}

func (ks *KVServer) Stop() {
	ks.raftServer.Stop()
	ks.raftLog.Close()
	ks.stateMachine.Close()
	ks.stateStorage.Close()
	logger.Info("Key-Value Stopped")
}

func (ks *KVServer) Put(_ context.Context, req *PutRequest) (*PutResponse, error) {
	if ks.raftServer.State() == raft.State_Learner {
		return &PutResponse{Status: ResponseStatus_Learner}, nil
	}
	cmd := NewInsertCommandFromKV(req.Key, req.Value)
	offerReplyFuture := ks.raftServer.OfferCommand(cmd)
	offerReply, err := offerReplyFuture.Get()
	if err != nil {
		return nil, err
	}
	offerR := offerReply.(*raft.OfferResponse)
	res := &PutResponse{
		Status:   ResponseStatus(offerR.Status.Number()),
		LeaderId: offerR.LeaderId,
	}
	return res, err
}

func (ks *KVServer) Delete(_ context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	if ks.raftServer.State() == raft.State_Learner {
		return &DeleteResponse{Status: ResponseStatus_Learner}, nil
	}
	cmd := NewDeleteCommandFromKey(req.Key)
	offerReplyFuture := ks.raftServer.OfferCommand(cmd)
	offerReply, err := offerReplyFuture.Get()
	if err != nil {
		return nil, err
	}
	offerR := offerReply.(*raft.OfferResponse)
	res := &DeleteResponse{
		Status:   ResponseStatus(offerR.Status.Number()),
		LeaderId: offerR.LeaderId,
	}
	return res, err
}

func (ks *KVServer) Get(_ context.Context, req *GetRequest) (*GetResponse, error) {
	if ks.raftServer.State() == raft.State_Learner {
		return &GetResponse{Status: ResponseStatus_Learner}, nil
	}
	res := &GetResponse{
		Key:    req.Key,
		Value:  ks.stateMachine.Get(req.Key),
		Status: ResponseStatus_Success,
	}
	return res, nil
}
