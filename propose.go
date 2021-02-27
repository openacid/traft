package traft

// request sent to Loop() to propose a cmd
type proposeReq struct {
	cmd   *Cmd
	finCh chan *ProposeReply
}

func (tr *TRaft) hdlPropose(cmd *Cmd, finCh chan<- *ProposeReply) {
	id := tr.Id
	me := tr.Status[id]
	now := uSecondI64()

	if now > me.VoteExpireAt {
		lg.Infow("hdl-propose:VoteExpired", "me.VoteExpireAt-now", me.VoteExpireAt-now)
		// no valid leader for now
		finCh <- &ProposeReply{
			OK:  false,
			Err: "vote expired",
		}
		lg.Infow("hdl-propose: returning")
		return
	}

	if me.VotedFor.Id != id {
		finCh <- &ProposeReply{
			OK:          false,
			Err:         "I am not leader",
			OtherLeader: me.VotedFor.Clone(),
		}
		return
	}

	rec := tr.AddLog(cmd)
	lg.Infow("hdl-propose:added-rec", "rec", rec.ShortStr(), "rec.Overrides:", rec.Overrides.DebugStr())

	me.Accepted.Union(rec.Overrides)

	go tr.forwardLog(
		me.VotedFor.Clone(),
		tr.Config.Clone(),
		[]*LogRecord{rec},
		func(rst *logForwardRst) {
			if rst.err != nil {
				finCh <- &ProposeReply{
					OK:  false,
					Err: rst.err.Error(),
				}
			} else {
				finCh <- &ProposeReply{
					OK: true,
				}
			}
		})
}
