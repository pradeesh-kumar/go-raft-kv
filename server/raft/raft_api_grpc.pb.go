// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: raft/raft_api.proto

package raft

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	RaftProtocolService_RequestVote_FullMethodName        = "/raft.RaftProtocolService/RequestVote"
	RaftProtocolService_AppendEntries_FullMethodName      = "/raft.RaftProtocolService/AppendEntries"
	RaftProtocolService_TransferLeadership_FullMethodName = "/raft.RaftProtocolService/TransferLeadership"
	RaftProtocolService_TimeoutNow_FullMethodName         = "/raft.RaftProtocolService/TimeoutNow"
	RaftProtocolService_AddServer_FullMethodName          = "/raft.RaftProtocolService/AddServer"
	RaftProtocolService_RemoveServer_FullMethodName       = "/raft.RaftProtocolService/RemoveServer"
	RaftProtocolService_InstallSnapshot_FullMethodName    = "/raft.RaftProtocolService/InstallSnapshot"
)

// RaftProtocolServiceClient is the client API for RaftProtocolService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RaftProtocolServiceClient interface {
	// Election and Log Replication
	RequestVote(ctx context.Context, in *VoteRequest, opts ...grpc.CallOption) (*VoteResponse, error)
	AppendEntries(ctx context.Context, in *AppendEntriesRequest, opts ...grpc.CallOption) (*AppendEntriesResponse, error)
	// Leadership Transfer
	TransferLeadership(ctx context.Context, in *TransferLeadershipRequest, opts ...grpc.CallOption) (*TransferLeadershipResponse, error)
	TimeoutNow(ctx context.Context, in *TimeoutNowRequest, opts ...grpc.CallOption) (*TimeoutNowResponse, error)
	// Cluster Membership Change
	AddServer(ctx context.Context, in *AddServerRequest, opts ...grpc.CallOption) (*AddServerResponse, error)
	RemoveServer(ctx context.Context, in *RemoveServerRequest, opts ...grpc.CallOption) (*RemoveServerResponse, error)
	// Snapshot
	InstallSnapshot(ctx context.Context, in *InstallSnapshotRequest, opts ...grpc.CallOption) (*InstallSnapshotResponse, error)
}

type raftProtocolServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRaftProtocolServiceClient(cc grpc.ClientConnInterface) RaftProtocolServiceClient {
	return &raftProtocolServiceClient{cc}
}

func (c *raftProtocolServiceClient) RequestVote(ctx context.Context, in *VoteRequest, opts ...grpc.CallOption) (*VoteResponse, error) {
	out := new(VoteResponse)
	err := c.cc.Invoke(ctx, RaftProtocolService_RequestVote_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftProtocolServiceClient) AppendEntries(ctx context.Context, in *AppendEntriesRequest, opts ...grpc.CallOption) (*AppendEntriesResponse, error) {
	out := new(AppendEntriesResponse)
	err := c.cc.Invoke(ctx, RaftProtocolService_AppendEntries_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftProtocolServiceClient) TransferLeadership(ctx context.Context, in *TransferLeadershipRequest, opts ...grpc.CallOption) (*TransferLeadershipResponse, error) {
	out := new(TransferLeadershipResponse)
	err := c.cc.Invoke(ctx, RaftProtocolService_TransferLeadership_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftProtocolServiceClient) TimeoutNow(ctx context.Context, in *TimeoutNowRequest, opts ...grpc.CallOption) (*TimeoutNowResponse, error) {
	out := new(TimeoutNowResponse)
	err := c.cc.Invoke(ctx, RaftProtocolService_TimeoutNow_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftProtocolServiceClient) AddServer(ctx context.Context, in *AddServerRequest, opts ...grpc.CallOption) (*AddServerResponse, error) {
	out := new(AddServerResponse)
	err := c.cc.Invoke(ctx, RaftProtocolService_AddServer_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftProtocolServiceClient) RemoveServer(ctx context.Context, in *RemoveServerRequest, opts ...grpc.CallOption) (*RemoveServerResponse, error) {
	out := new(RemoveServerResponse)
	err := c.cc.Invoke(ctx, RaftProtocolService_RemoveServer_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftProtocolServiceClient) InstallSnapshot(ctx context.Context, in *InstallSnapshotRequest, opts ...grpc.CallOption) (*InstallSnapshotResponse, error) {
	out := new(InstallSnapshotResponse)
	err := c.cc.Invoke(ctx, RaftProtocolService_InstallSnapshot_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RaftProtocolServiceServer is the server API for RaftProtocolService service.
// All implementations must embed UnimplementedRaftProtocolServiceServer
// for forward compatibility
type RaftProtocolServiceServer interface {
	// Election and Log Replication
	RequestVote(context.Context, *VoteRequest) (*VoteResponse, error)
	AppendEntries(context.Context, *AppendEntriesRequest) (*AppendEntriesResponse, error)
	// Leadership Transfer
	TransferLeadership(context.Context, *TransferLeadershipRequest) (*TransferLeadershipResponse, error)
	TimeoutNow(context.Context, *TimeoutNowRequest) (*TimeoutNowResponse, error)
	// Cluster Membership Change
	AddServer(context.Context, *AddServerRequest) (*AddServerResponse, error)
	RemoveServer(context.Context, *RemoveServerRequest) (*RemoveServerResponse, error)
	// Snapshot
	InstallSnapshot(context.Context, *InstallSnapshotRequest) (*InstallSnapshotResponse, error)
	mustEmbedUnimplementedRaftProtocolServiceServer()
}

// UnimplementedRaftProtocolServiceServer must be embedded to have forward compatible implementations.
type UnimplementedRaftProtocolServiceServer struct {
}

func (UnimplementedRaftProtocolServiceServer) RequestVote(context.Context, *VoteRequest) (*VoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVote not implemented")
}
func (UnimplementedRaftProtocolServiceServer) AppendEntries(context.Context, *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendEntries not implemented")
}
func (UnimplementedRaftProtocolServiceServer) TransferLeadership(context.Context, *TransferLeadershipRequest) (*TransferLeadershipResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TransferLeadership not implemented")
}
func (UnimplementedRaftProtocolServiceServer) TimeoutNow(context.Context, *TimeoutNowRequest) (*TimeoutNowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TimeoutNow not implemented")
}
func (UnimplementedRaftProtocolServiceServer) AddServer(context.Context, *AddServerRequest) (*AddServerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddServer not implemented")
}
func (UnimplementedRaftProtocolServiceServer) RemoveServer(context.Context, *RemoveServerRequest) (*RemoveServerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveServer not implemented")
}
func (UnimplementedRaftProtocolServiceServer) InstallSnapshot(context.Context, *InstallSnapshotRequest) (*InstallSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InstallSnapshot not implemented")
}
func (UnimplementedRaftProtocolServiceServer) mustEmbedUnimplementedRaftProtocolServiceServer() {}

// UnsafeRaftProtocolServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RaftProtocolServiceServer will
// result in compilation errors.
type UnsafeRaftProtocolServiceServer interface {
	mustEmbedUnimplementedRaftProtocolServiceServer()
}

func RegisterRaftProtocolServiceServer(s grpc.ServiceRegistrar, srv RaftProtocolServiceServer) {
	s.RegisterService(&RaftProtocolService_ServiceDesc, srv)
}

func _RaftProtocolService_RequestVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftProtocolServiceServer).RequestVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftProtocolService_RequestVote_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftProtocolServiceServer).RequestVote(ctx, req.(*VoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftProtocolService_AppendEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppendEntriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftProtocolServiceServer).AppendEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftProtocolService_AppendEntries_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftProtocolServiceServer).AppendEntries(ctx, req.(*AppendEntriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftProtocolService_TransferLeadership_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransferLeadershipRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftProtocolServiceServer).TransferLeadership(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftProtocolService_TransferLeadership_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftProtocolServiceServer).TransferLeadership(ctx, req.(*TransferLeadershipRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftProtocolService_TimeoutNow_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TimeoutNowRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftProtocolServiceServer).TimeoutNow(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftProtocolService_TimeoutNow_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftProtocolServiceServer).TimeoutNow(ctx, req.(*TimeoutNowRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftProtocolService_AddServer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddServerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftProtocolServiceServer).AddServer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftProtocolService_AddServer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftProtocolServiceServer).AddServer(ctx, req.(*AddServerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftProtocolService_RemoveServer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveServerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftProtocolServiceServer).RemoveServer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftProtocolService_RemoveServer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftProtocolServiceServer).RemoveServer(ctx, req.(*RemoveServerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftProtocolService_InstallSnapshot_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InstallSnapshotRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftProtocolServiceServer).InstallSnapshot(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftProtocolService_InstallSnapshot_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftProtocolServiceServer).InstallSnapshot(ctx, req.(*InstallSnapshotRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RaftProtocolService_ServiceDesc is the grpc.ServiceDesc for RaftProtocolService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RaftProtocolService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "raft.RaftProtocolService",
	HandlerType: (*RaftProtocolServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestVote",
			Handler:    _RaftProtocolService_RequestVote_Handler,
		},
		{
			MethodName: "AppendEntries",
			Handler:    _RaftProtocolService_AppendEntries_Handler,
		},
		{
			MethodName: "TransferLeadership",
			Handler:    _RaftProtocolService_TransferLeadership_Handler,
		},
		{
			MethodName: "TimeoutNow",
			Handler:    _RaftProtocolService_TimeoutNow_Handler,
		},
		{
			MethodName: "AddServer",
			Handler:    _RaftProtocolService_AddServer_Handler,
		},
		{
			MethodName: "RemoveServer",
			Handler:    _RaftProtocolService_RemoveServer_Handler,
		},
		{
			MethodName: "InstallSnapshot",
			Handler:    _RaftProtocolService_InstallSnapshot_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "raft/raft_api.proto",
}
