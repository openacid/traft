package traft

import context "context"

func (tr *TRaft) Vote(ctx context.Context, req *VoteReq) (*VoteReply, error) {
	repl := query(tr.actionCh, "vote", req).v.(*VoteReply)
	return repl, nil
}

func (tr *TRaft) Replicate(ctx context.Context, req *ReplicateReq) (*ReplicateReply, error) {

	rst := query(tr.actionCh, "replicate", req)
	return rst.v.(*ReplicateReply), nil

	// TODO: if a newer committer is seen, non-committed logs
	// can be sure to stale and should be cleaned.
}

func (tr *TRaft) Propose(ctx context.Context, cmd *Cmd) (*ProposeReply, error) {

	rst := query(tr.actionCh, "propose", cmd)
	var otherLeader *LeaderId
	if rst.v != nil {
		otherLeader = rst.v.(*LeaderId)
	}
	return &ProposeReply{
		OK:          rst.ok,
		OtherLeader: otherLeader,
	}, nil
}
