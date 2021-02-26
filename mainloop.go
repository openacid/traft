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

// query the mainloop goroutine for something, by other goroutines, such as
// update traft state or get some info.
func (tr *TRaft) query(arg interface{}) *queryRst {
	rstCh := make(chan *queryRst)
	tr.actionCh <- &queryBody{arg, rstCh}
	rst := <-rstCh
	lg.Infow("chan-query",
		// "arg", arg,
		"rst.err", rst.err,
		"rst.v", toStr(rst.v))
	return rst
}

// Loop handles actions from other components.
// This is the only goroutine that is allowed to update traft state.
// Any info to send out of this goroutine must be cloned.
func (tr *TRaft) Loop() {

	for {
		select {
		case <-tr.shutdown:
			return
		case a := <-tr.actionCh:

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

// checkStatus checks if TRaft status violate consistency requirement.
// This is just a routine for debug.
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
