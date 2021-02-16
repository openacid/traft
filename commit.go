package traft

import "github.com/pkg/errors"

// the request sending to Loop() to commit a continuous range of logs
type commitReq struct {
	committer *LeaderId
	lsns      []int64
}

func (tr *TRaft) leaderUpdateCommitted(
	committer *LeaderId, lsns []int64) error {

	id := tr.Id
	me := tr.Status[id]
	if committer.Equal(me.VotedFor) {
		// Logs are intact.
		// may be expired, Logs wont change unless VotedFor changes.

		// NOTE: using start, end lsn to describe logs requires every
		// forwarding action operates on continous logs
		for i := lsns[0]; i < lsns[1]; i++ {
			r := tr.Logs[i-tr.LogOffset]
			me.Committed.Union(r.Overrides)
		}

		return nil
	}

	err := errors.Wrapf(ErrLeaderLost,
		"committer: %s, current %s",
		committer.ShortStr(), me.VotedFor.ShortStr(),
	)
	lg.Infow("leaderUpdateCommitted", "err", err)
	return err
}
