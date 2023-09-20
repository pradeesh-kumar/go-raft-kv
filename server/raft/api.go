package raft

import (
	"errors"
	"time"
)

const largeFuture = time.Hour * 24 * 365

var ErrUnknownCommand = errors.New("unknown command")

type RaftServer interface {
	Start()
	Stop()
	OfferCommand(command Command) Future
}

type RaftState interface {
	start()
	name() State
	handleRPC(*RPC)
	stop()
}

type StateMachine interface {
	Apply(log []*Record)
	Close()
}

type StateStorage interface {
	Store(*PersistentState) error
	Retrieve() (*PersistentState, error)
}

type RaftLog interface {
	Append(*Record) (uint64, error)
	Read(offset uint64) (*Record, error)
	ReadAllSince(offset uint64) ([]*Record, error)
	ReadBatchSince(offset uint64, batchSize int) ([]*Record, error)
	Truncate(lowest uint64) error
	LowestOffset() (uint64, error)
	HighestOffset() (uint64, error)
}

type Command interface {
	Bytes() []byte
}

type Transport interface {
	RegisterMessageHandler(messageHandler RPCHandler)
	Start()
	Stop()
	BroadcastVote(BroadcastRequest[*VoteRequest]) BroadcastResponse[*VoteResponse]
	SendAppendEntries(Payload[*AppendEntriesRequest]) (Payload[*AppendEntriesResponse], error)
	SendTimeoutRequest(Payload[*TimeoutNowRequest]) (Payload[*TimeoutNowResponse], error)
}

type RPCHandler interface {
	HandleRPC(message any) Future
}

type BroadcastRequest[REQ any] []Payload[REQ]
type BroadcastResponse[RES any] []Payload[RES]

type Payload[M any] struct {
	ServerId
	ServerAddress
	Message M
}

type Node struct {
	id      ServerId
	address ServerAddress
}

type Follower struct {
	Node
	nextIndex  uint64
	matchIndex uint64
}

func NewPayload[M any](sid ServerId, address ServerAddress, message M) Payload[M] {
	return Payload[M]{
		ServerId:      sid,
		ServerAddress: address,
		Message:       message,
	}
}
