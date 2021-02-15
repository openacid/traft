package traft

import context "context"

func (tr *TRaft) Vote(ctx context.Context, req *VoteReq) (*VoteReply, error) {
	rst := query(tr.actionCh, "vote", req)
	return rst.v.(*VoteReply), nil
}

func (tr *TRaft) Replicate(ctx context.Context, req *ReplicateReq) (*ReplicateReply, error) {

	// TODO: if a newer committer is seen, non-committed logs
	// can be sure to stale and should be cleaned.

	rst := query(tr.actionCh, "replicate", req)
	return rst.v.(*ReplicateReply), nil
}

func (tr *TRaft) Propose(ctx context.Context, cmd *Cmd) (*ProposeReply, error) {
	rst := query(tr.actionCh, "propose", cmd)
	return rst.v.(*ProposeReply), nil
}
