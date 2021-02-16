package traft

import context "context"

func (tr *TRaft) Vote(ctx context.Context, req *VoteReq) (*VoteReply, error) {
	rst := query(tr.actionCh, "vote", req)
	return rst.v.(*VoteReply), nil
}

func (tr *TRaft) LogForward(ctx context.Context, req *LogForwardReq) (*LogForwardReply, error) {

	// TODO: if a newer committer is seen, non-committed logs
	// can be sure to stale and should be cleaned.

	rst := query(tr.actionCh, "replicate", req)
	return rst.v.(*LogForwardReply), nil
}

func (tr *TRaft) Propose(ctx context.Context, cmd *Cmd) (*ProposeReply, error) {
	finCh := make(chan *ProposeReply, 1)
	query(tr.actionCh, "propose", &proposeReq{cmd, finCh})

	lg.Infow("waitingFor:finCh")
	rst := <-finCh
	lg.Infow("got:finCh", "rst", rst)
	return rst, nil
}
