package raft

type CandidateState struct {
	raftServer *RaftServerImpl
}

func newCandidateState(r *RaftServerImpl) RaftState {
	return &CandidateState{r}
}

func (s *CandidateState) start() {

}

func (*CandidateState) name() State {
	return State_Candidate
}

func (s *CandidateState) handleRPC(rpc *RPC) {
	switch cmd := rpc.cmd.(type) {
	case *VoteRequest:
		rpc.reply.Set(s.raftServer.processVoteRequest(cmd))
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
		if cmd.Term > s.raftServer.currentTerm {
			s.raftServer.changeState(newFollowerState(s.raftServer))
		} else {
			rpc.reply.Set(&InstallSnapshotResponse{Term: s.raftServer.currentTerm}, nil)
		}
	default:
		rpc.reply.Set(nil, ErrUnknownCommand)
	}
}

func (s *CandidateState) offerCommand(requests []*OfferRequest) {
	response := &OfferResponse{
		Status:   ResponseStatus_NotLeader,
		LeaderId: string(s.raftServer.currentLeader),
	}
	for _, r := range requests {
		r.reply.Set(response, nil)
	}
}

func (*CandidateState) stop() {
}
