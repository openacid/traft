package traft

type queryRst struct {
	v   interface{}
	ok  bool
	err error
}

type queryBody struct {
	operation string
	arg       interface{}
	rstCh     chan *queryRst
}

// For other goroutine to ask mainloop to query
func query(queryCh chan<- *queryBody, operation string, arg interface{}) *queryRst {
	rstCh := make(chan *queryRst)
	queryCh <- &queryBody{operation, arg, rstCh}
	rst := <-rstCh
	lg.Infow("chan-query",
		"operation", operation,
		"arg", arg,
		"rst.ok", rst.ok,
		"rst.v", toStr(rst.v))
	return rst
}

// Loop handles actions from other components.
func (tr *TRaft) Loop() {

	shutdown := tr.shutdown
	act := tr.actionCh

	id := tr.Id

	for {
		select {
		case <-shutdown:
			return
		case a := <-act:
			switch a.operation {
			case "logStat":
				a.rstCh <- &queryRst{
					v: ExportLogStatus(tr.Status[id]),
				}
			case "leaderStat":
				lg.Infow("send-leader", "v", ExportLeaderStatus(tr.Status[id]))
				a.rstCh <- &queryRst{
					v: ExportLeaderStatus(tr.Status[id]),
				}
			case "config":
				a.rstCh <- &queryRst{
					v: tr.Config.Clone(),
				}
			case "vote":
				a.rstCh <- &queryRst{
					v: tr.hdlVoteReq(a.arg.(*VoteReq)),
				}
			case "set_voted":
				leadst := a.arg.(*LeaderStatus)
				me := tr.Status[id]
				if leadst.VotedFor.Cmp(me.VotedFor) >= 0 {
					me.VotedFor = leadst.VotedFor.Clone()
					me.VoteExpireAt = leadst.VoteExpireAt
					a.rstCh <- &queryRst{ok: true}
				} else {
					a.rstCh <- &queryRst{ok: false}
				}
			case "update_leaderAndLog":
				// TODO rename this operation

				lal := a.arg.(*leaderAndVotes)
				leadst := lal.leaderStat
				votes := lal.votes

				me := tr.Status[id]

				if leadst.VotedFor.Cmp(me.VotedFor) == 0 {
					me.VotedFor = leadst.VotedFor.Clone()
					me.VoteExpireAt = leadst.VoteExpireAt

					tr.internalMergeLogs(votes)
					// TODO update Committer to this replica
					// then going on replicating these logs to others.
					//
					// TODO update local view of status of other replicas.
					for _, v := range votes {
						if v.Committer.Equal(me.Committer) {
							tr.Status[v.Id].Accepted = v.Accepted.Clone()
						} else {
							// if committers are different, the leader can no be
							// sure whether a follower has identical logs
							tr.Status[v.Id].Accepted = v.Committed.Clone()
						}
						tr.Status[v.Id].Committed = v.Committed.Clone()

						tr.Status[v.Id].Committer = v.Committer.Clone()
					}
					me.Committer = leadst.VotedFor.Clone()
					a.rstCh <- &queryRst{ok: true}
				} else {
					a.rstCh <- &queryRst{ok: false}
				}

			case "propose":
				p := a.arg.(*proposeReq)
				tr.hdlPropose(p.cmd, p.finCh)
				a.rstCh <- &queryRst{}

			case "replicate":
				// receive logs forwarded from leader
				rreq := a.arg.(*LogForwardReq)
				reply := tr.hdlLogForward(rreq)
				a.rstCh <- &queryRst{
					ok: reply.OK,
					v:  reply,
				}

			case "commit":
				c := a.arg.(*commitReq)
				err := tr.leaderUpdateCommitted(
					c.committer, c.lsns,
				)
				if err == nil {
					a.rstCh <- &queryRst{
						ok: true,
					}

				} else {
					a.rstCh <- &queryRst{
						ok: false,
					}
				}

			case "func":
				f := a.arg.(func() error)
				err := f()
				a.rstCh <- &queryRst{err: err}
			default:
				panic("unknown action" + a.operation)
			}
		}
	}
}
