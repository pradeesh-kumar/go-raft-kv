package kv

import (
	"context"
	"fmt"
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"time"
)

type TransportOpts struct {
	broadcastTimeout time.Duration
	address          raft.ServerAddress
}

type GrpcTransport struct {
	server         *grpc.Server
	nodeClientMap  map[raft.ServerId]*raft.RaftProtocolServiceClient
	messageHandler raft.RPCHandler
	TransportOpts
	raft.UnimplementedRaftProtocolServiceServer
	raft.UnimplementedRaftClientServiceServer
}

func NewGrpcTransport(transportOpts TransportOpts) *GrpcTransport {
	transport := &GrpcTransport{
		TransportOpts: transportOpts,
		server:        grpc.NewServer(),
		nodeClientMap: make(map[raft.ServerId]*raft.RaftProtocolServiceClient),
	}
	return transport
}

func (t *GrpcTransport) RegisterMessageHandler(messageHandler raft.RPCHandler) {
	t.messageHandler = messageHandler
}

func (t *GrpcTransport) RegisterClientService(serviceDesc *grpc.ServiceDesc, methodHandler any) {
	t.server.RegisterService(serviceDesc, methodHandler)
}

func (t *GrpcTransport) Start() {
	go func() {
		listener, err := net.Listen("tcp", string(t.address))
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("Listener started on ", t.address)
		if err = t.server.Serve(listener); err != nil {
			logger.Fatal(err)
		}
	}()
	// Block until server starts
	time.Sleep(5 * time.Second)
}

func dispatch[REQ any, RES any](t *GrpcTransport, req REQ) (*RES, error) {
	val, err := t.messageHandler.HandleRPC(req).Get()
	if val == nil {
		return nil, err
	}
	return val.(*RES), err
}

func (t *GrpcTransport) RequestVote(c context.Context, req *raft.VoteRequest) (*raft.VoteResponse, error) {
	return dispatch[*raft.VoteRequest, raft.VoteResponse](t, req)
}

func (t *GrpcTransport) AppendEntries(c context.Context, req *raft.AppendEntriesRequest) (*raft.AppendEntriesResponse, error) {
	return dispatch[*raft.AppendEntriesRequest, raft.AppendEntriesResponse](t, req)
}

func (t *GrpcTransport) TransferLeadership(c context.Context, req *raft.TransferLeadershipRequest) (*raft.TransferLeadershipResponse, error) {
	return dispatch[*raft.TransferLeadershipRequest, raft.TransferLeadershipResponse](t, req)
}

func (t *GrpcTransport) TimeoutNow(c context.Context, req *raft.TimeoutNowRequest) (*raft.TimeoutNowResponse, error) {
	return dispatch[*raft.TimeoutNowRequest, raft.TimeoutNowResponse](t, req)
}

func (t *GrpcTransport) AddServer(c context.Context, req *raft.AddServerRequest) (*raft.AddServerResponse, error) {
	return dispatch[*raft.AddServerRequest, raft.AddServerResponse](t, req)
}
func (t *GrpcTransport) RemoveServer(c context.Context, req *raft.RemoveServerRequest) (*raft.RemoveServerResponse, error) {
	return dispatch[*raft.RemoveServerRequest, raft.RemoveServerResponse](t, req)
}

func (t *GrpcTransport) SendTimeoutRequest(req raft.Payload[*raft.TimeoutNowRequest]) (raft.Payload[*raft.TimeoutNowResponse], error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.broadcastTimeout)
	defer cancel()
	var client raft.RaftProtocolServiceClient
	if _, ok := t.nodeClientMap[req.ServerId]; !ok {
		client = createClient(req.ServerAddress)
		t.nodeClientMap[req.ServerId] = &client
	}
	client = *t.nodeClientMap[req.ServerId]
	resPayload := raft.NewPayload[*raft.TimeoutNowResponse](req.ServerId, req.ServerAddress, nil)
	if res, err := client.TimeoutNow(ctx, req.Message); err != nil {
		return resPayload, err
	} else {
		resPayload.Message = res
		return resPayload, nil
	}
}

func (t *GrpcTransport) SendAppendEntries(req raft.Payload[*raft.AppendEntriesRequest]) (raft.Payload[*raft.AppendEntriesResponse], error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.broadcastTimeout)
	defer cancel()
	var client raft.RaftProtocolServiceClient
	if _, ok := t.nodeClientMap[req.ServerId]; !ok {
		client = createClient(req.ServerAddress)
		t.nodeClientMap[req.ServerId] = &client
	}
	client = *t.nodeClientMap[req.ServerId]
	resPayload := raft.NewPayload[*raft.AppendEntriesResponse](req.ServerId, req.ServerAddress, nil)
	if res, err := client.AppendEntries(ctx, req.Message); err != nil {
		return resPayload, err
	} else {
		resPayload.Message = res
		return resPayload, nil
	}
}

type VoteFunction func(ctx context.Context, in *raft.VoteRequest, opts ...grpc.CallOption) (*raft.VoteResponse, error)

func (t *GrpcTransport) BroadcastVote(request raft.BroadcastRequest[*raft.VoteRequest]) raft.BroadcastResponse[*raft.VoteResponse] {
	grpcMethodSupplier := func(c raft.RaftProtocolServiceClient) func(ctx context.Context, in *raft.VoteRequest, opts ...grpc.CallOption) (*raft.VoteResponse, error) {
		return c.RequestVote
	}
	return broadcast(request, t.broadcastTimeout, t.nodeClientMap, grpcMethodSupplier)
}

func broadcast[REQ any, RES any](
	request raft.BroadcastRequest[*REQ],
	timeout time.Duration,
	nodeClientMap map[raft.ServerId]*raft.RaftProtocolServiceClient,
	grpcMethodSupplier func(raft.RaftProtocolServiceClient) func(ctx context.Context, in *REQ, opts ...grpc.CallOption) (*RES, error),
) raft.BroadcastResponse[*RES] {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resChan, errChan := make(chan raft.Payload[*RES], len(request)), make(chan error)
	for _, nodePayload := range request {
		var client raft.RaftProtocolServiceClient
		if _, ok := nodeClientMap[nodePayload.ServerId]; !ok {
			client = createClient(nodePayload.ServerAddress)
			nodeClientMap[nodePayload.ServerId] = &client
		}
		client = *nodeClientMap[nodePayload.ServerId]
		go func(p raft.Payload[*REQ]) {
			res, err := grpcMethodSupplier(client)(ctx, p.Message)
			if deadline, _ := ctx.Deadline(); time.Now().Before(deadline) {
				if err != nil {
					errChan <- fmt.Errorf("failed to broadcast to the node %s %v", p.ServerId, err)
				} else {
					resChan <- raft.NewPayload(p.ServerId, p.ServerAddress, res)
				}
			}
		}(nodePayload)
	}
	return collectBroadcastResult(ctx, resChan, errChan, len(request))
}

func collectBroadcastResult[T any](ctx context.Context, resChan chan raft.Payload[T], errChan chan error, resultCount int) raft.BroadcastResponse[T] {
	defer close(resChan)
	defer close(errChan)
	var broadcastResponse raft.BroadcastResponse[T]
	for i := 1; i <= resultCount; i++ {
		select {
		case res := <-resChan:
			broadcastResponse = append(broadcastResponse, res)
		case err := <-errChan:
			logger.Error(err)
		case <-ctx.Done():
			logger.Warn("Broadcast Context timed out!")
			goto label
		}
	}
label:
	return broadcastResponse
}

func createClient(addr raft.ServerAddress) raft.RaftProtocolServiceClient {
	conn, err := grpc.Dial(string(addr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Error while creating client connection for the address: ", addr)
		return nil
	}
	return raft.NewRaftProtocolServiceClient(conn)
}

type ClientProtocolServer struct {
	address raft.ServerAddress
	server  *grpc.Server
	UnimplementedClientProtocolServiceServer
}

func (t *GrpcTransport) Stop() {
	t.server.GracefulStop()
}
