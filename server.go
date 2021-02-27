package traft

// TRaftServer impl

import (
	context "context"
	fmt "fmt"
	"math/rand"
	"time"

	"github.com/openacid/low/mathext/util"
	"github.com/pkg/errors"
	// "google.golang.org/protobuf/proto"
)

var leaderLease = int64(time.Second * 1)

// init a TRaft for test, all logs are `set x=lsn`
func (tr *TRaft) initTraft(
	// proposer of the logs
	committer *LeaderId,
	// author of the logs
	author *LeaderId,
	// log seq numbers to generate.
	lsns []int64,
	nilLogs map[int64]bool,
	committed []int64,
	votedFor *LeaderId,
) {
	id := tr.Id

	tr.LogOffset, tr.Logs = buildPseudoLogs(author, lsns, nilLogs)

	tr.Status[id].Committer = committer.Clone()
	tr.Status[id].Accepted = NewTailBitmap(0, lsns...)

	if committed == nil {
		tr.Status[id].Committed = NewTailBitmap(0)
	} else {
		tr.Status[id].Committed = NewTailBitmap(0, committed...)
	}

	tr.Status[id].VotedFor = votedFor.Clone()

	tr.checkStatus()
}

func buildPseudoLogs(
	// author of the logs
	author *LeaderId,
	// log seq numbers to generate.
	lsns []int64,
	nilLogs map[int64]bool,
) (int64, []*Record) {
	logs := make([]*Record, 0)
	if len(lsns) == 0 {
		return 0, logs
	}

	last := lsns[len(lsns)-1]
	start := lsns[0]
	for i := start; i <= last; i++ {
		logs = append(logs, &Record{})
	}

	for _, lsn := range lsns {
		if nilLogs != nil && nilLogs[lsn] {
		} else {
			logs[lsn-start] = NewRecord(
				author.Clone(),
				lsn,
				NewCmdI64("set", "x", lsn))
		}
	}
	return start, logs
}

// a func for test purpose only
func (tr *TRaft) addlogs(cmds ...interface{}) {
	me := tr.Status[tr.Id]
	for _, cs := range cmds {
		cmd := toCmd(cs)
		r := tr.AddLog(cmd)
		me.Accepted.Union(r.Overrides)
	}
}

// establishLeadership updates leader state when a election approved by a quorum.
func (tr *TRaft) establishLeadership(currVote *LeaderId, replies []*VoteReply) {

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
func (tr *TRaft) mergeFollowerLogs(votes []*VoteReply) {

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
	lg.Infow("mergeFollowerLogs", "maxCommitter", maxCommitter)

	l := me.Accepted.Len()
	for lsn := me.Accepted.Offset; lsn < l; lsn++ {
		if me.Accepted.Get(lsn) != 0 {
			continue
		}

		var rec *Record
		for _, vr := range chosen {
			r := vr.PopRecord(lsn)
			fmt.Println("lsn:", lsn, "r:", r.ShortStr())
			if r == nil {
				continue
			}

			if rec != nil && rec.Author.Cmp(r.Author) != 0 {
				panic("wtf")
			}

			rec = r
			// TODO if r is not nil: break
		}

		// TODO fill in with empty log
		if rec == nil {
			continue
		}

		tr.Logs[lsn-tr.LogOffset] = rec
		me.Accepted.Set(lsn)

		lg.Infow("merge-log",
			"lsn", lsn,
			"committer", maxCommitter,
			"record", rec)
	}
}

// chooseMaxCommitterReplies chooses the max Committer and the vote-replies with the max Committer.
// logs with Committer smaller than me are discarded too.
func (tr *TRaft) chooseMaxCommitterReplies(replies []*VoteReply) (*LeaderId, []*VoteReply) {
	me := tr.Status[tr.Id]
	maxCommitter := me.Committer
	for _, v := range replies {
		if v.Committer.Cmp(maxCommitter) > 0 {
			maxCommitter = v.Committer
		}
	}
	chosen := make([]*VoteReply, 0, len(replies))
	for _, v := range replies {
		if v.Committer.Cmp(maxCommitter) == 0 {
			chosen = append(chosen, v)
		}
	}
	return maxCommitter, chosen
}

// run forever to elect itself as leader if there is no leader in this cluster.
func (tr *TRaft) VoteLoop() {

	id := tr.Id

	slp := tr.sleep

	maxStaleTermSleep := time.Millisecond * 200
	heartBeatSleep := time.Millisecond * 200
	followerSleep := time.Millisecond * 200

	for tr.running {
		var currVote *LeaderId
		var expireAt int64
		var logst *LogStatus
		var config *ClusterConfig

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

		voteReplies, err, higher := VoteOnce(
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
// VoteReply-s: if vote granted by a quorum, returns collected replies.
//				Otherwise returns nil.
// error: ErrStaleLog, ErrStaleTermId, ErrTimeout.
// higherTerm: if seen, upgrade term and retry
func VoteOnce(
	candidate *LeaderId,
	logStatus *LogStatus,
	config *ClusterConfig,
) ([]*VoteReply, error, int64) {

	// TODO vote need cluster id:
	// a stale member may try to elect on another cluster.

	id := candidate.Id

	replies := make([]*VoteReply, 0)

	req := &VoteReq{
		Candidate: candidate,
		Committer: logStatus.GetCommitter(),
		Accepted:  logStatus.GetAccepted(),
	}

	type voteRst struct {
		from  *ReplicaInfo
		reply *VoteReply
		err   error
	}

	ch := make(chan *voteRst)

	for _, rinfo := range config.Members {
		if rinfo.Id == id {
			continue
		}

		go func(rinfo ReplicaInfo, ch chan *voteRst) {
			rpcTo(rinfo.Addr, func(cli TRaftClient, ctx context.Context) {

				// lg.Infow("grpc-send-req", "addr", rinfo.Addr)
				reply, err := cli.Vote(ctx, req)
				// lg.Infow("grpc-recv-reply", "from", rinfo, "reply", reply, "err", err)

				ch <- &voteRst{&rinfo, reply, err}
			})
		}(*rinfo, ch)
	}

	received := uint64(0)
	// I vote myself
	received |= 1 << uint(config.Members[id].Position)
	higherTerm := int64(-1)
	var logErr error
	waitingFor := len(config.Members) - 1

	for waitingFor > 0 {
		select {
		case res := <-ch:

			lg.Infow("vote-once:got-reply", "reply", res.reply, "err", res.err)

			if res.err != nil {
				waitingFor--
				continue
			}

			repl := res.reply

			if repl.VotedFor.Equal(candidate) {
				// vote granted
				replies = append(replies, repl)

				received |= 1 << uint(res.from.Position)
				if config.IsQuorum(received) {
					// TODO cancel timer
					return replies, nil, -1
				}
			} else {
				if repl.VotedFor.Cmp(candidate) > 0 {
					higherTerm = util.MaxI64(higherTerm, repl.VotedFor.Term)
				}

				if CmpLogStatus(repl, logStatus) > 0 {
					// TODO cancel timer
					logErr = errors.Wrapf(ErrStaleLog,
						"local: committer:%s max-lsn:%d remote: committer:%s max-lsn:%d",
						logStatus.GetCommitter().ShortStr(),
						logStatus.GetAccepted().Len(),
						repl.Committer.ShortStr(),
						repl.Accepted.Len())
				}
			}

			waitingFor--

		case <-time.After(time.Second):
			// timeout
			// TODO cancel timer
			return nil, errors.Wrapf(ErrTimeout, "voting"), higherTerm
		}
	}

	if logErr != nil {
		return nil, logErr, higherTerm
	}

	err := errors.Wrapf(ErrStaleTermId, "seen a higher term:%d", higherTerm)
	return nil, err, higherTerm
}

// Only a established leader should use this func.
// no lock protection, must be called from Loop()
func (tr *TRaft) AddLog(cmd *Cmd) *Record {

	me := tr.Status[tr.Id]

	if me.VotedFor.Id != tr.Id {
		panic("wtf")
	}

	lsn := tr.LogOffset + int64(len(tr.Logs))

	r := NewRecord(me.VotedFor, lsn, cmd)

	// find the first interfering record.

	var i int
	for i = len(tr.Logs) - 1; i >= 0; i-- {
		prev := tr.Logs[i]
		if r.Interfering(prev) {
			r.Overrides = prev.Overrides.Clone()
			break
		}
	}

	if i == -1 {
		// there is not a interfering record.
		r.Overrides = NewTailBitmap(0)
	}

	r.Overrides.Set(lsn)

	// all log I do not know must be executed in order.
	// Because I do not know of the intefering relations.
	r.Depends = NewTailBitmap(tr.LogOffset)

	// reduce bitmap size by removing unknown logs
	r.Overrides.Union(NewTailBitmap(tr.LogOffset & ^63))

	tr.Logs = append(tr.Logs, r)

	return r
}

// no lock protect, must be called by TRaft.Loop()
func (tr *TRaft) hdlVoteReq(req *VoteReq) *VoteReply {

	id := tr.Id

	me := tr.Status[tr.Id]

	// A vote reply just send back a voter's status.
	// It is the candidate's responsibility to check if a voter granted it.
	repl := &VoteReply{
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
	repl.VotedFor = req.Candidate.Clone()

	// send back the logs I have but the candidate does not.

	logs := make([]*Record, 0)

	start := me.Accepted.Offset
	end := me.Accepted.Len()
	for i := start; i < end; i++ {
		if me.Accepted.Get(i) != 0 && req.Accepted.Get(i) == 0 {
			logs = append(logs, tr.Logs[i-tr.LogOffset])
		}
	}

	repl.Logs = logs

	return repl
}
