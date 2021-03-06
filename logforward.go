package traft

import (
	"time"

	"github.com/pkg/errors"
)

// the result of forwarding logs from leader to follower
type logForwardRst struct {
	from  *ReplicaInfo
	reply *LogForwardReply
	err   error
}

// forward log from leader to follower concurrently
func (tr *TRaft) forwardLog(
	committer *LeaderId,
	config *Cluster,
	logs []*LogRecord,
	callback func(*logForwardRst),
) {

	lsns := []int64{logs[0].Seq, logs[len(logs)-1].Seq + 1}
	lg.Infow("forward", "LSNs", lsns, "cmtr", committer)

	req := &LogForwardReq{
		Committer: committer,
		Logs:      logs,
	}

	id := tr.Id

	// TODO
	timeout := time.Second
	sess := rpcToAll(id, config, meth.LogForward, req, timeout)

	for res := range sess.resCh {

		lg.Infow("logforward:recv-reply", "res", res)

		if sess.updateOKBitmap(res) {

			rst := tr.query(func() error {
				return tr.leaderUpdateCommitted(
					committer, lsns,
				)
			})

			if rst.err == nil {
				lg.Infow("forward:a-quorum-done")
				callback(&logForwardRst{})
			} else {
				// TODO let the root cause to generate the error
				callback(&logForwardRst{
					err: errors.Wrapf(rst.err, "forward"),
				})
			}
			// LogForward does not cancel, try best to send logs to followers.
			return
		}
	}

	lg.Infow("forward:timeout", "cmtr", committer.ShortStr())
	callback(&logForwardRst{
		err: errors.Wrapf(ErrTimeout, "forward"),
	})
}

// hdlLogForward handles LogForward request on a follower
// LogForward is similar to paxos-phase-2.
func (tr *TRaft) hdlLogForward(req *LogForwardReq) *LogForwardReply {
	me := tr.Status[tr.Id]
	now := uSecondI64()

	lg.Infow("hdl-logforward", "req", req)
	lg.Infow("hdl-logforward", "me", me)

	cr := req.Committer.Cmp(me.VotedFor)

	// If req.Committer > me.VotedFor, it is a valid leader too.
	// It is safe to accept its log.
	// This is a common optimization of paxos: an Acceptor accepts request if rnd >= lastrnd.
	// See: https://blog.openacid.com/algo/paxos/#slide-42

	if cr < 0 && now < me.VoteExpireAt {
		lg.Infow("hdl-logforward: illegal committer",
			"req.Commiter", req.Committer,
			"me.VotedFor", me.VotedFor,
			"me.VoteExpireAt-now", me.VoteExpireAt-now)

		return &LogForwardReply{
			OK:       false,
			VotedFor: me.VotedFor.Clone(),
		}
	}

	if cr > 0 {
		me.VotedFor = req.Committer.Clone()
		me.VoteExpireAt = now + leaderLease
	}

	// TODO apply req.Committed

	cr = req.Committer.Cmp(me.Committer)
	if cr > 0 {
		lg.Infow("hdl-log-forward: newer committer",
			"req.Committer", req.Committer,
			"me.Committer", me.Committer,
		)

		// if req.Committer is newer, discard all non-committed logs
		// Because non-committed local log may have been overridden by some new leader.
		me.Accepted = me.Committed.Clone()

		i := len(tr.Logs) - 1
		for ; i >= 0; i-- {
			r := tr.Logs[i]
			if r.Empty() {
				continue
			}

			if me.Accepted.Get(r.Seq) == 0 {
				tr.Logs[i] = &LogRecord{}
			}
		}
	}

	// add new logs

	for _, r := range req.Logs {
		lsn := r.Seq
		idx := lsn - tr.LogOffset

		for int(idx) >= len(tr.Logs) {
			tr.Logs = append(tr.Logs, &LogRecord{})
		}

		if me.Accepted.Get(lsn) != 0 {
			if !tr.Logs[idx].Empty() && !tr.Logs[idx].Equal(r) {
				panic("wtf")
			}
		}
		tr.Logs[idx] = r

		me.Accepted.Union(r.Overrides)

		lg.Infow("hdl-logforward", "accept-log", r)
		lg.Infow("hdl-logforward", "accepted", me.Accepted)
	}

	// TODO refine me
	// remove empty logs at top
	for len(tr.Logs) > 0 {
		l := len(tr.Logs)
		if tr.Logs[l-1].Empty() {
			tr.Logs = tr.Logs[:l-1]
		} else {
			break
		}
	}

	me.Committer = req.Committer.Clone()

	me.UpdatedCommitted(req.Committer, req.Committed)

	return &LogForwardReply{
		OK:        true,
		VotedFor:  me.VotedFor.Clone(),
		Accepted:  me.Accepted.Clone(),
		Committed: me.Committed.Clone(),
	}
}
