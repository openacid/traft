package traft

import fmt "fmt"

type queryBody struct {
	arg   interface{}
	rstCh chan *queryRst
}

type queryRst struct {
	v   interface{}
	err error
}

// For other goroutine to ask mainloop to query
func (tr *TRaft) query(arg interface{}) *queryRst {
	rstCh := make(chan *queryRst)
	tr.actionCh <- &queryBody{arg, rstCh}
	rst := <-rstCh
	lg.Infow("chan-query",
		"arg", arg,
		"rst.err", rst.err,
		"rst.v", toStr(rst.v))
	return rst
}

// Loop handles actions from other components.
func (tr *TRaft) Loop() {

	shutdown := tr.shutdown
	act := tr.actionCh

	for {
		select {
		case <-shutdown:
			return
		case a := <-act:

			tr.checkStatus()

			switch f := a.arg.(type) {
			case func() error:
				err := f()
				a.rstCh <- &queryRst{err: err}
			case func() interface{}:
				v := f()
				a.rstCh <- &queryRst{v: v}
			default:
				panic("unknown func signature:" + fmt.Sprintf("%v", a.arg))
			}

			tr.checkStatus()
		}
	}
}

// check if TRaft status violate consistency requirement.
func (tr *TRaft) checkStatus() {
	id := tr.Id
	me := tr.Status[id]

	// committer can never greater than voted leader
	if me.Committer.Cmp(me.VotedFor) > 0 {
		panic(
			fmt.Sprintf("Commiter > VotedFor: Id:%d %s %s",
				id,
				me.Committer.ShortStr(),
				me.VotedFor.ShortStr(),
			))
	}
}
