package traft

import fmt "fmt"

type queryRst struct {
	v   interface{}
	err error
}

type queryBody struct {
	operation string
	arg       interface{}
	rstCh     chan *queryRst
}

// For other goroutine to ask mainloop to query
func (tr *TRaft) query(operation string, arg interface{}) *queryRst {
	rstCh := make(chan *queryRst)
	tr.actionCh <- &queryBody{operation, arg, rstCh}
	rst := <-rstCh
	lg.Infow("chan-query",
		"operation", operation,
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

			switch a.operation {
			case "func":
				f := a.arg.(func() error)
				err := f()
				a.rstCh <- &queryRst{err: err}
			case "funcv":
				f := a.arg.(func() interface{})
				v := f()
				a.rstCh <- &queryRst{v: v}
			default:
				panic("unknown action" + a.operation)
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
