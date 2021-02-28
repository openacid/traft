package traft

import "context"

func (tr *TRaft) Elect(ctx context.Context, req *ElectReq) (*ElectReply, error) {
	var reply *ElectReply
	rst := tr.query( func() error {
		reply= tr.hdlElectReq(req)
		return nil
	})
	_ = rst
	return reply,nil
}

func (tr *TRaft) LogForward(ctx context.Context, req *LogForwardReq) (*LogForwardReply, error) {

	// TODO: if a newer committer is seen, non-committed logs
	// can be sure to stale and should be cleaned.

	var reply *LogForwardReply
	rst := tr.query( func() error {
		reply = tr.hdlLogForward(req)
		return nil
	})
	_ = rst
	return reply, nil
}

func (tr *TRaft) Propose(ctx context.Context, cmd *Cmd) (*ProposeReply, error) {

	finCh := make(chan *ProposeReply, 1)

	rst := tr.query( func() error {
		tr.hdlPropose(cmd, finCh)
		return nil
	})
	_ = rst

	lg.Infow("waitingFor:finCh")
	reply := <-finCh
	lg.Infow("got:finCh", "reply", reply)

	return reply, nil
}
