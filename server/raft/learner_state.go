package raft

import "os"

type LearnerState struct {
	raftServer          *RaftServerImpl
	currentSnapshotFile *os.File
}

func newLearnerState(r *RaftServerImpl) RaftState {
	return &LearnerState{raftServer: r, currentSnapshotFile: nil}
}

func (s *LearnerState) start() {

}

func (*LearnerState) name() State {
	return State_Candidate
}

func (s *LearnerState) handleRPC(rpc *RPC) {
	switch cmd := rpc.cmd.(type) {
	case *VoteRequest:
		// Do not vote - Learner is a non-voting member
		rpc.reply.Set(&VoteResponse{Term: s.raftServer.currentTerm, VoteGranted: false}, nil)
	case *AppendEntriesRequest:
		rpc.reply.Set(s.raftServer.processAppendEntriesRequest(cmd))
	case *TransferLeadershipRequest:
		response := &TransferLeadershipResponse{
			Success:  false,
			LeaderId: string(s.raftServer.currentLeader),
		}
		rpc.reply.Set(response, nil)
	case *TimeoutNowRequest:
		rpc.reply.Set(&TimeoutNowResponse{Success: false}, nil)
	case *AddServerRequest:
		rpc.reply.Set(&AddServerResponse{Status: ResponseStatus_NotLeader, LeaderId: string(s.raftServer.currentLeader)}, nil)
	case *RemoveServerRequest:
		rpc.reply.Set(&RemoveServerResponse{Status: ResponseStatus_NotLeader, LeaderId: string(s.raftServer.currentLeader)}, nil)
	case []*OfferRequest:
		s.offerCommand(cmd)
	case *InstallSnapshotRequest:
		rpc.reply.Set(s.raftServer.processInstallSnapshotRequest(cmd))
	default:
		rpc.reply.Set(nil, ErrUnknownCommand)
	}
}

func (s *LearnerState) offerCommand(requests []*OfferRequest) {
	response := &OfferResponse{
		Status:   ResponseStatus_NotLeader,
		LeaderId: string(s.raftServer.currentLeader),
	}
	for _, r := range requests {
		r.reply.Set(response, nil)
	}
}

func (*LearnerState) stop() {
}
