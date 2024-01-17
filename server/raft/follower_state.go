package raft

type FollowerState struct {
	raftServer *RaftServerImpl
}

func newFollowerState(r *RaftServerImpl) RaftState {
	return &FollowerState{r}
}

func (s *FollowerState) start() {
}

func (*FollowerState) name() State {
	return State_Follower
}

func (s *FollowerState) handleRPC(rpc *RPC) {
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
		rpc.reply.Set(s.timeoutNow(cmd))
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

func (s *FollowerState) timeoutNow(req *TimeoutNowRequest) (*TimeoutNowResponse, error) {
	response := &TimeoutNowResponse{Success: false}
	if ServerId(req.LeaderId) != s.raftServer.currentLeader {
		return response, nil
	}
	s.raftServer.startElection()
	response.Success = s.raftServer.state.name() == State_Leader
	return response, nil
}

func (s *FollowerState) offerCommand(requests []*OfferRequest) {
	response := &OfferResponse{
		Status:   ResponseStatus_NotLeader,
		LeaderId: string(s.raftServer.currentLeader),
	}
	for _, r := range requests {
		r.reply.Set(response, nil)
	}
}

func (*FollowerState) stop() {
}
