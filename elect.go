package traft

import (
	"math/rand"
	"time"

	"github.com/openacid/low/mathext/util"
	"github.com/pkg/errors"
)

// run forever to elect itself as leader if there is no leader in this cluster.
func (tr *TRaft) ElectLoop() {

	id := tr.Id

	slp := tr.sleep

	maxStaleTermSleep := time.Millisecond * 200
	heartBeatSleep := time.Millisecond * 200
	followerSleep := time.Millisecond * 200

	for tr.running {
		var currVote *LeaderId
		var expireAt int64
		var logst *LogStatus
		var config *Cluster

		now := uSecondI64()
		lg.Infow("vote loop round start:",
			"Id", tr.Id,
		)

		err := tr.query(func() error {
			me := tr.Status[tr.Id]

			currVote = me.VotedFor.Clone()
			expireAt = me.VoteExpireAt
			logst = ExportLogStatus(tr.Status[tr.Id])
			config = tr.Config.Clone()

			if now < expireAt {
				return nil
			}

			// init state for voting myself

			me.VotedFor.Term++
			me.VotedFor.Id = tr.Id
			currVote = me.VotedFor.Clone()

			me.VoteExpireAt = uSecondI64() + leaderLease

			return errors.Wrapf(ErrNeedElect, "expireAt-now: %d", expireAt-now)

		}).err

		if err == nil {
			// TODO refine this: wait until VoteExpireAt and watch for missing
			// heartbeat.
			lg.Infow("leader-not-expired",
				"Id", tr.Id,
				"VotedFor", currVote,
				"leadst.VoteExpireAt-now", expireAt-now)

			if currVote.Id == tr.Id {
				// I am a leader
				// TODO heartbeat other replicas to keep leadership
				slp(heartBeatSleep)
			} else {
				slp(followerSleep)
			}
			continue
		}

		// call for a new leader!!!
		lg.Infow("leader-expired",
			"Id", tr.Id,
			"VotedFor", currVote,
			"leadst.VoteExpireAt-now", expireAt-now)

		tr.sendMsg("vote-start", currVote.ShortStr(), logst)

		voteReplies, err, higher := ElectOnce(
			currVote,
			logst,
			config,
		)

		lg.Infow("vote-loop:result", "Id", tr.Id, "voteReplies", voteReplies, "err", err, "higher", higher)

		if voteReplies == nil {
			// fail to elect me.
			tr.sendMsg("vote-fail", "err", err)
			tr.query(func() error {

				me := tr.Status[tr.Id]

				if currVote.Cmp(me.VotedFor) == 0 {
					// I did not vote other ones yet, and I am not leader.
					// reset it.
					me.VoteExpireAt = 0
				}

				return nil
			})

			// wait for some time by err
			switch errors.Cause(err) {
			case ErrStaleTermId:
				slp(time.Millisecond*5 + time.Duration(rand.Int63n(int64(maxStaleTermSleep))))
			case ErrTimeout:
				slp(time.Millisecond * 10)
			case ErrStaleLog:
				// I can not be the leader.
				// sleep a day. waiting for others to elect to be a leader.
				slp(time.Second * 86400)
			}
			continue
		}

		// granted by a quorum

		lg.Infow("to-update-leader", "votedFor", currVote)

		updateErr := tr.query(func() error {

			me := tr.Status[tr.Id]

			if currVote.Cmp(me.VotedFor) != 0 {
				return errors.Wrapf(ErrLeaderLost, "when updating leadership and follower state")
			}

			tr.establishLeadership(currVote, voteReplies)
			return nil

		}).err

		if updateErr != nil {
			tr.sendMsg("vote-fail", "reason:fail-to-update", currVote)
			lg.Infow("vote-fail", "Id", id,
				"currVote", currVote,
				"err", updateErr.Error(),
			)
			continue
		}

		tr.sendMsg("vote-win", currVote)
		slp(heartBeatSleep)
	}
}

// returns:
// ElectReply-s: if vote granted by a quorum, returns collected replies.
//				Otherwise returns nil.
// error: ErrStaleLog, ErrStaleTermId, ErrTimeout.
// higherTerm: if seen, upgrade term and retry
func ElectOnce(
	candidate *LeaderId,
	logStatus *LogStatus,
	cluster *Cluster,
) ([]*ElectReply, error, int64) {

	// TODO vote need cluster id:
	// a stale member may try to elect on another cluster.

	id := candidate.Id

	replies := make([]*ElectReply, 0)

	req := &ElectReq{
		Candidate: candidate,
		Committer: logStatus.GetCommitter(),
		Accepted:  logStatus.GetAccepted(),
	}

	type voteRst struct {
		from  *ReplicaInfo
		reply *ElectReply
		err   error
	}

	higherTerm := int64(-1)
	var logErr error

	timeout := time.Second
	sess := rpcToAll(id, cluster, meth.Elect, req, timeout)

	for res := range sess.resCh {

		reply := res.reply.(*ElectReply)
		lg.Infow("elect:recv-reply", "reply", reply, "res.err", res.err)

		if reply.OK {
			replies = append(replies, reply)
			if sess.updateOKBitmap(res) {
				// do not cancel
				return replies, nil, -1
			}
			continue
		}

		if reply.VotedFor.Cmp(candidate) > 0 {
			higherTerm = util.MaxI64(higherTerm, reply.VotedFor.Term)
		}

		if CmpLogStatus(reply, logStatus) > 0 {
			// TODO cancel timer
			logErr = errors.Wrapf(ErrStaleLog,
				"local: committer:%s max-lsn:%d remote: committer:%s max-lsn:%d",
				logStatus.GetCommitter().ShortStr(),
				logStatus.GetAccepted().Len(),
				reply.Committer.ShortStr(),
				reply.Accepted.Len())
		}
	}

	if logErr != nil {
		return nil, logErr, higherTerm
	}

	err := errors.Wrapf(ErrStaleTermId, "seen a higher term:%d", higherTerm)
	return nil, err, higherTerm
}

// no lock protect, must be called by TRaft.Loop()
func (tr *TRaft) hdlElectReq(req *ElectReq) *ElectReply {

	id := tr.Id

	me := tr.Status[tr.Id]

	// A vote reply just send back a voter's status.
	// It is the candidate's responsibility to check if a voter granted it.
	repl := &ElectReply{
		OK:        false,
		Id:        id,
		VotedFor:  me.VotedFor.Clone(),
		Committer: me.Committer.Clone(),
		Accepted:  me.Accepted.Clone(),
		Committed: me.Committed.Clone(),
	}

	lg.Infow("handleVoteReq",
		"Id", id,
		"req.Candidate", req.Candidate,
		"me.Committer", me.Committer.ShortStr(),
		"me.Accepted", me.Accepted.ShortStr(),
		"me.VotedFor", me.VotedFor.ShortStr(),
		"req.Committer", req.Committer.ShortStr(),
		"req.Accepted", req.Accepted.ShortStr(),
	)

	if CmpLogStatus(req, me) < 0 {
		// I have more logs than the candidate.
		// It cant be a leader.
		tr.sendMsg("hdl-vote-req:reject-by-logstat",
			"req.Candidate", req.Candidate,
			"me.Committer", me.Committer,
			"me.Accepted", me.Accepted,
			"req.Committer", req.Committer,
			"req.Accepted", req.Accepted,
		)
		return repl
	}

	// candidate has the upto date logs.

	r := req.Candidate.Cmp(me.VotedFor)
	if r < 0 {
		// I've voted for other leader with higher privilege.
		// This candidate could not be a legal leader.
		// just send back enssential info to info it.
		tr.sendMsg("hdl-vote-req:reject-by-term-id",
			"req.Candidate", req.Candidate,
			"me.VotedFor", me.VotedFor,
		)
		return repl
	}

	// grant vote

	lg.Infow("voted", "id", id, "VotedFor", me.VotedFor)
	tr.sendMsg("hdl-vote-req:grant",
		"req.Candidate", req.Candidate,
		"me.VotedFor", me.VotedFor)

	me.VotedFor = req.Candidate.Clone()
	me.VoteExpireAt = uSecondI64() + leaderLease
	repl.OK = true
	repl.VotedFor = req.Candidate.Clone()

	// send back the logs I have but the candidate does not.

	logs := make([]*LogRecord, 0)

	lg.Infow("hdlElectReq", "me.Accepted", me.Accepted)
	lg.Infow("hdlElectReq", "req.Accepted", req.Accepted)
	start := me.Accepted.Offset
	end := me.Accepted.Len()
	for i := start; i < end; i++ {
		if me.Accepted.Get(i) != 0 && req.Accepted.Get(i) == 0 {
			r := tr.Logs[i-tr.LogOffset]
			logs = append(logs, r)
			lg.Infow("hdlElectReq:send-log", "r", r)
		}
	}

	repl.Logs = logs

	return repl
}

// establishLeadership updates leader state when a election approved by a quorum.
func (tr *TRaft) establishLeadership(currVote *LeaderId, replies []*ElectReply) {

	me := tr.Status[tr.Id]

	// not to update expire time.
	// let the leader expire earlier than follower to reduce chance that follower reject replication from leader.

	tr.mergeFollowerLogs(replies)

	// then going on replicating these logs to others.
	//
	// TODO update local view of status of other replicas.
	for _, r := range replies {
		follower := tr.Status[r.Id]
		if r.Committer.Equal(me.Committer) {
			follower.Accepted = r.Accepted.Clone()
		} else {
			// if committers are different, the leader can no be
			// sure whether a follower has identical logs
			follower.Accepted = r.Committed.Clone()
		}
		follower.Committed = r.Committed.Clone()

		follower.Committer = r.Committer.Clone()
	}

	// Leader accept all the logs it sees
	me.Committer = currVote.Clone()

}

// find the max committer log to fill in local log holes.
func (tr *TRaft) mergeFollowerLogs(votes []*ElectReply) {

	// TODO if the leader chose Logs[i] from replica `r`, e.g. R[r].Logs[i]
	// then the logs R[r].Logs[:i] are safe to choose.
	// Because if a different R[r'].Logs[j] is committed, for a j <= i
	// the leader that written R[r].Log[i] must have chosen R[r'].Logs[j] .
	// ∴ R[r].Logs[j] == R[r'].Logs[j]
	//
	// For now 2021 Feb 14,
	// we just erase all non-committed logs on followers.

	id := tr.Id
	me := tr.Status[id]

	maxCommitter, chosen := tr.chooseMaxCommitterReplies(votes)
	lg.Infow("mergeFollowerLogs", "maxCommitter", maxCommitter, "chosen", chosen)
	lg.Infow("mergeFollowerLogs", "mylogs", RecordsShortStr(tr.Logs))

	l := me.Accepted.Len()
	for lsn := me.Accepted.Offset; lsn < l; lsn++ {
		if me.Accepted.Get(lsn) != 0 {
			continue
		}

		rec := getLog(lsn, chosen)

		// TODO fill in with empty log
		if rec.Empty() {
			continue
		}

		tr.Logs[lsn-tr.LogOffset] = rec
		me.Accepted.Union(rec.Overrides)

		lg.Infow("merge-log",
			"lsn", lsn,
			"committer", maxCommitter,
			"record", rec)
	}
}

// getLog returns one log record if a log  with the specified lsn presents in any vote replies.
func getLog(lsn int64, replies []*ElectReply) *LogRecord {
	var rec *LogRecord
	for _, vr := range replies {
		r := vr.PopRecord(lsn)
		if r == nil {
			continue
		}

		if rec != nil && rec.Author.Cmp(r.Author) != 0 {
			panic("wtf")
		}

		rec = r
		// TODO if r is not nil: break
	}

	return rec
}

// chooseMaxCommitterReplies chooses the max Committer and the vote-replies with the max Committer.
// logs with Committer smaller than me are discarded too.
func (tr *TRaft) chooseMaxCommitterReplies(replies []*ElectReply) (*LeaderId, []*ElectReply) {
	me := tr.Status[tr.Id]
	maxCommitter := me.Committer
	for _, v := range replies {
		if v.Committer.Cmp(maxCommitter) > 0 {
			maxCommitter = v.Committer
		}
	}
	chosen := make([]*ElectReply, 0, len(replies))
	for _, v := range replies {
		if v.Committer.Cmp(maxCommitter) == 0 {
			chosen = append(chosen, v)
		}
	}
	return maxCommitter, chosen
}
