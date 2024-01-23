// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v4.25.2
// source: raft/client_api.proto

package raft

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type RegisterClientRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RegisterClientRequest) Reset() {
	*x = RegisterClientRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_raft_client_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterClientRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterClientRequest) ProtoMessage() {}

func (x *RegisterClientRequest) ProtoReflect() protoreflect.Message {
	mi := &file_raft_client_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterClientRequest.ProtoReflect.Descriptor instead.
func (*RegisterClientRequest) Descriptor() ([]byte, []int) {
	return file_raft_client_api_proto_rawDescGZIP(), []int{0}
}

type RegisterClientResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status   ResponseStatus `protobuf:"varint,1,opt,name=status,proto3,enum=raft.ResponseStatus" json:"status,omitempty"`
	ClientId int32          `protobuf:"varint,2,opt,name=clientId,proto3" json:"clientId,omitempty"`
	LeaderId string         `protobuf:"bytes,3,opt,name=leaderId,proto3" json:"leaderId,omitempty"`
}

func (x *RegisterClientResponse) Reset() {
	*x = RegisterClientResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_raft_client_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterClientResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterClientResponse) ProtoMessage() {}

func (x *RegisterClientResponse) ProtoReflect() protoreflect.Message {
	mi := &file_raft_client_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterClientResponse.ProtoReflect.Descriptor instead.
func (*RegisterClientResponse) Descriptor() ([]byte, []int) {
	return file_raft_client_api_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterClientResponse) GetStatus() ResponseStatus {
	if x != nil {
		return x.Status
	}
	return ResponseStatus_Success
}

func (x *RegisterClientResponse) GetClientId() int32 {
	if x != nil {
		return x.ClientId
	}
	return 0
}

func (x *RegisterClientResponse) GetLeaderId() string {
	if x != nil {
		return x.LeaderId
	}
	return ""
}

type WriteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId       int32  `protobuf:"varint,1,opt,name=clientId,proto3" json:"clientId,omitempty"`
	SequenceNumber int32  `protobuf:"varint,2,opt,name=sequenceNumber,proto3" json:"sequenceNumber,omitempty"`
	Command        []byte `protobuf:"bytes,3,opt,name=command,proto3" json:"command,omitempty"`
}

func (x *WriteRequest) Reset() {
	*x = WriteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_raft_client_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteRequest) ProtoMessage() {}

func (x *WriteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_raft_client_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteRequest.ProtoReflect.Descriptor instead.
func (*WriteRequest) Descriptor() ([]byte, []int) {
	return file_raft_client_api_proto_rawDescGZIP(), []int{2}
}

func (x *WriteRequest) GetClientId() int32 {
	if x != nil {
		return x.ClientId
	}
	return 0
}

func (x *WriteRequest) GetSequenceNumber() int32 {
	if x != nil {
		return x.SequenceNumber
	}
	return 0
}

func (x *WriteRequest) GetCommand() []byte {
	if x != nil {
		return x.Command
	}
	return nil
}

type WriteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status   ResponseStatus `protobuf:"varint,1,opt,name=status,proto3,enum=raft.ResponseStatus" json:"status,omitempty"`
	Response []byte         `protobuf:"bytes,2,opt,name=response,proto3" json:"response,omitempty"`
	LeaderId string         `protobuf:"bytes,3,opt,name=leaderId,proto3" json:"leaderId,omitempty"`
}

func (x *WriteResponse) Reset() {
	*x = WriteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_raft_client_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteResponse) ProtoMessage() {}

func (x *WriteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_raft_client_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteResponse.ProtoReflect.Descriptor instead.
func (*WriteResponse) Descriptor() ([]byte, []int) {
	return file_raft_client_api_proto_rawDescGZIP(), []int{3}
}

func (x *WriteResponse) GetStatus() ResponseStatus {
	if x != nil {
		return x.Status
	}
	return ResponseStatus_Success
}

func (x *WriteResponse) GetResponse() []byte {
	if x != nil {
		return x.Response
	}
	return nil
}

func (x *WriteResponse) GetLeaderId() string {
	if x != nil {
		return x.LeaderId
	}
	return ""
}

type QueryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Query []byte `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
}

func (x *QueryRequest) Reset() {
	*x = QueryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_raft_client_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryRequest) ProtoMessage() {}

func (x *QueryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_raft_client_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryRequest.ProtoReflect.Descriptor instead.
func (*QueryRequest) Descriptor() ([]byte, []int) {
	return file_raft_client_api_proto_rawDescGZIP(), []int{4}
}

func (x *QueryRequest) GetQuery() []byte {
	if x != nil {
		return x.Query
	}
	return nil
}

type QueryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status   ResponseStatus `protobuf:"varint,1,opt,name=status,proto3,enum=raft.ResponseStatus" json:"status,omitempty"`
	Response []byte         `protobuf:"bytes,2,opt,name=response,proto3" json:"response,omitempty"`
	LeaderId string         `protobuf:"bytes,3,opt,name=leaderId,proto3" json:"leaderId,omitempty"`
}

func (x *QueryResponse) Reset() {
	*x = QueryResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_raft_client_api_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryResponse) ProtoMessage() {}

func (x *QueryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_raft_client_api_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryResponse.ProtoReflect.Descriptor instead.
func (*QueryResponse) Descriptor() ([]byte, []int) {
	return file_raft_client_api_proto_rawDescGZIP(), []int{5}
}

func (x *QueryResponse) GetStatus() ResponseStatus {
	if x != nil {
		return x.Status
	}
	return ResponseStatus_Success
}

func (x *QueryResponse) GetResponse() []byte {
	if x != nil {
		return x.Response
	}
	return nil
}

func (x *QueryResponse) GetLeaderId() string {
	if x != nil {
		return x.LeaderId
	}
	return ""
}

var File_raft_client_api_proto protoreflect.FileDescriptor

var file_raft_client_api_proto_rawDesc = []byte{
	0x0a, 0x15, 0x72, 0x61, 0x66, 0x74, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x61, 0x70,
	0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x72, 0x61, 0x66, 0x74, 0x1a, 0x13, 0x72,
	0x61, 0x66, 0x74, 0x2f, 0x72, 0x61, 0x66, 0x74, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x17, 0x0a, 0x15, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x43, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x7e, 0x0a, 0x16, 0x52,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x2e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12,
	0x1a, 0x0a, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x22, 0x6c, 0x0a, 0x0c, 0x57,
	0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x26, 0x0a, 0x0e, 0x73, 0x65, 0x71, 0x75, 0x65,
	0x6e, 0x63, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0e, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12,
	0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x22, 0x75, 0x0a, 0x0d, 0x57, 0x72, 0x69,
	0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x72, 0x61, 0x66,
	0x74, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64,
	0x22, 0x24, 0x0a, 0x0c, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x22, 0x75, 0x0a, 0x0d, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x2e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x32, 0xca, 0x01,
	0x0a, 0x11, 0x52, 0x61, 0x66, 0x74, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x4d, 0x0a, 0x0e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x43,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x1b, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x2e, 0x52, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x65, 0x72, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x65, 0x72, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x32, 0x0a, 0x05, 0x57, 0x72, 0x69, 0x74, 0x65, 0x12, 0x12, 0x2e, 0x72, 0x61,
	0x66, 0x74, 0x2e, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x13, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x2e, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x32, 0x0a, 0x05, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12,
	0x12, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x61, 0x64, 0x65, 0x65, 0x73,
	0x68, 0x2d, 0x6b, 0x75, 0x6d, 0x61, 0x72, 0x2f, 0x67, 0x6f, 0x2d, 0x72, 0x61, 0x66, 0x74, 0x2d,
	0x6b, 0x76, 0x2f, 0x72, 0x61, 0x66, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_raft_client_api_proto_rawDescOnce sync.Once
	file_raft_client_api_proto_rawDescData = file_raft_client_api_proto_rawDesc
)

func file_raft_client_api_proto_rawDescGZIP() []byte {
	file_raft_client_api_proto_rawDescOnce.Do(func() {
		file_raft_client_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_raft_client_api_proto_rawDescData)
	})
	return file_raft_client_api_proto_rawDescData
}

var file_raft_client_api_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_raft_client_api_proto_goTypes = []interface{}{
	(*RegisterClientRequest)(nil),  // 0: raft.RegisterClientRequest
	(*RegisterClientResponse)(nil), // 1: raft.RegisterClientResponse
	(*WriteRequest)(nil),           // 2: raft.WriteRequest
	(*WriteResponse)(nil),          // 3: raft.WriteResponse
	(*QueryRequest)(nil),           // 4: raft.QueryRequest
	(*QueryResponse)(nil),          // 5: raft.QueryResponse
	(ResponseStatus)(0),            // 6: raft.ResponseStatus
}
var file_raft_client_api_proto_depIdxs = []int32{
	6, // 0: raft.RegisterClientResponse.status:type_name -> raft.ResponseStatus
	6, // 1: raft.WriteResponse.status:type_name -> raft.ResponseStatus
	6, // 2: raft.QueryResponse.status:type_name -> raft.ResponseStatus
	0, // 3: raft.RaftClientService.RegisterClient:input_type -> raft.RegisterClientRequest
	2, // 4: raft.RaftClientService.Write:input_type -> raft.WriteRequest
	4, // 5: raft.RaftClientService.Query:input_type -> raft.QueryRequest
	1, // 6: raft.RaftClientService.RegisterClient:output_type -> raft.RegisterClientResponse
	3, // 7: raft.RaftClientService.Write:output_type -> raft.WriteResponse
	5, // 8: raft.RaftClientService.Query:output_type -> raft.QueryResponse
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_raft_client_api_proto_init() }
func file_raft_client_api_proto_init() {
	if File_raft_client_api_proto != nil {
		return
	}
	file_raft_raft_api_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_raft_client_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterClientRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_raft_client_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterClientResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_raft_client_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_raft_client_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_raft_client_api_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_raft_client_api_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_raft_client_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_raft_client_api_proto_goTypes,
		DependencyIndexes: file_raft_client_api_proto_depIdxs,
		MessageInfos:      file_raft_client_api_proto_msgTypes,
	}.Build()
	File_raft_client_api_proto = out.File
	file_raft_client_api_proto_rawDesc = nil
	file_raft_client_api_proto_goTypes = nil
	file_raft_client_api_proto_depIdxs = nil
}
