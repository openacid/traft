package traft

// TRaftServer impl

import (
	context "context"
	"encoding/json"
	fmt "fmt"
	"math/rand"
	sync "sync"
	"time"

	"github.com/openacid/low/mathext/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	// "google.golang.org/protobuf/proto"
)

var (
	ErrStaleLog    = errors.New("local log is stale")
	ErrStaleTermId = errors.New("local Term-Id is stale")
	ErrTimeout     = errors.New("timeout")
)

var (
	llg = zap.NewNop()
	lg  *zap.SugaredLogger
)

func init() {
	// if os.Getenv("CLUSTER_DEBUG") != "" {
	// }
	var err error
	// llg, err = zap.NewProduction()
	llg, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	lg = llg.Sugar()

	// initZap()
}

func initZap() {
	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["stdout", "/tmp/logs"],
	  "errorOutputPaths": ["stderr"],
	  "initialFields": {"foo": "bar"},
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}

	var err error
	llg, err = cfg.Build()
	if err != nil {
		panic(err)
	}
	defer llg.Sync()

	llg.Info("logger construction succeeded")

	lg = llg.Sugar()

}

type TRaft struct {
	// TODO lock first
	mu sync.Mutex

	stopped bool

	// close it to notify all goroutines to shutdown.
	shutdown chan struct{}

	// Communication channel with Loop().
	// Only Loop() modifies state of TRaft.
	// Other goroutines send an queryBody through this channel and wait for an
	// operation reply.
	actionCh chan *queryBody

	// for external component to receive traft state changes.
	MsgCh chan string

	grpcServer *grpc.Server

	wg sync.WaitGroup

	Node
}

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

type leaderAndVotes struct {
	leaderStat *LeaderStatus
	votes      []*VoteReply
}

type queryRst struct {
	logStat    *LogStatus
	leaderStat *LeaderStatus
	config     *ClusterConfig
	voteReply  *VoteReply

	v interface{}

	ok bool
}

type queryBody struct {
	operation string
	arg       interface{}
	rstCh     chan *queryRst
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
					v: tr.internalVote(a.arg.(*VoteReq)),
				}
			case "update_leaderStat":
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

				if leadst.VotedFor.Cmp(me.VotedFor) >= 0 {
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

						tr.Status[v.Id].Committer = v.Committer
					}
					me.Committer = leadst.VotedFor.Clone()
					a.rstCh <- &queryRst{ok: true}
				} else {
					a.rstCh <- &queryRst{ok: false}
				}
			}
		}
	}
}

// find the max committer log to fill in local log holes.
func (tr *TRaft) internalMergeLogs(votes []*VoteReply) {

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

	l := me.Accepted.Len()
	for i := me.Accepted.Offset; i < l; i++ {
		if me.Accepted.Get(i) != 0 {
			continue
		}

		maxCommitter := NewLeaderId(0, 0)
		var maxRec *Record
		// var isCommitted bool
		for _, vr := range votes {
			r := vr.PopRecord(i)
			fmt.Println("lsn:", i, "r:", r.ShortStr())
			if r == nil {
				continue
			}

			cmpRst := maxCommitter.Cmp(vr.Committer)
			if cmpRst == 0 {
				if !maxRec.Equal(r) {
					panic("wtf: same committer different log")
				}
			}

			if cmpRst < 0 {
				maxCommitter = vr.Committer
				maxRec = r
			}

			// if !isCommitted && vr.Committed.Get(i) != 0 {
			//     isCommitted = true
			// }
		}

		if maxRec == nil {
			continue
		}

		tr.Logs[i-tr.LogOffset] = maxRec
		me.Accepted.Set(i)
		// if isCommitted {
		//     me.Committed.Set(i)
		// }

		lg.Infow("merge-log",
			"lsn", i,
			"committer", maxCommitter,
			"record", maxRec)
	}
}

// For other goroutine to ask mainloop to query/update
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

var leaderLease = int64(time.Second * 1)

// run forever to elect itself as leader if there is no leader in this cluster.
func (tr *TRaft) VoteLoop() {

	shutdown := tr.shutdown
	act := tr.actionCh

	running := true
	id := tr.Id

	// return true if shutting down
	slp := func(sleep time.Duration) {
		select {
		case <-time.After(sleep):
			// lg.Infow("VoteLoop woke up")
		case <-shutdown:
			running = false
		}
	}

	// maxStaleTermSleep := time.Millisecond * 10
	// heartBeatSleep := time.Millisecond * 10
	// followerSleep := time.Millisecond * 10

	maxStaleTermSleep := time.Millisecond * 200
	heartBeatSleep := time.Millisecond * 200
	followerSleep := time.Millisecond * 200

	for running {
		leadst := query(act, "leaderStat", nil).v.(*LeaderStatus)

		now := uSecond()

		// TODO refine this: wait until VoteExpireAt and watch for missing
		// heartbeat.
		if now < leadst.VoteExpireAt {

			lg.Infow("leader-not-expired",
				"Id", tr.Id,
				"VotedFor", leadst.VotedFor,
				"leadst.VoteExpireAt-now", leadst.VoteExpireAt-now)

			if leadst.VotedFor.Id == tr.Id {
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
			"VotedFor", leadst.VotedFor,
			"leadst.VoteExpireAt-now", leadst.VoteExpireAt-now)

		logst := query(act, "logStat", nil).v.(*LogStatus)
		config := query(act, "config", nil).v.(*ClusterConfig)

		// vote myself
		leadst.VotedFor.Term++
		leadst.VotedFor.Id = tr.Id

		{
			// update local vote first
			ok := query(act, "update_leaderStat", leadst).ok
			if !ok {
				// voted for other replica
				leadst = query(act, "leaderStat", nil).v.(*LeaderStatus)
				lg.Infow("reload-leader",
					"Id", id,
					"leadst.VotedFor", leadst.VotedFor,
					"leadst.VoteExpireAt", leadst.VoteExpireAt,
				)
				continue
			}
		}

		tr.sendMsg("vote-start", leadst.VotedFor.ShortStr(), logst)

		voted, err, higher := VoteOnce(
			leadst.VotedFor,
			logst,
			config,
		)

		lg.Infow("vote result", "voted", voted, "err", err, "higher", higher)

		if voted != nil {
			// granted by a quorum

			leadst.VoteExpireAt = uSecond() + leaderLease

			lg.Infow("to-update-leader", "leadst", leadst.VoteExpireAt)

			ok := query(act, "update_leaderAndLog", &leaderAndVotes{
				leadst,
				voted,
			}).ok

			if ok {
				tr.sendMsg("vote-win", leadst)
				slp(heartBeatSleep)
			} else {
				tr.sendMsg("vote-fail", "reason:fail-to-update", leadst)
				lg.Infow("reload-leader",
					"Id", id,
					"leadst.VotedFor", leadst.VotedFor,
					"leadst.VoteExpireAt", leadst.VoteExpireAt,
				)
			}
			continue
		}

		tr.sendMsg("vote-fail", "err", err)

		// not voted

		switch errors.Cause(err) {
		case ErrStaleTermId:
			slp(time.Millisecond*5 + time.Duration(rand.Int63n(int64(maxStaleTermSleep))))
			// leadst.VotedFor.Term = higher + 1
		case ErrTimeout:
			slp(time.Millisecond * 10)
		case ErrStaleLog:
			// I can not be the leader.
			// sleep a day. waiting for others to elect to be a leader.
			slp(time.Second * 86400)
		}
	}
}

// returns:
// VoteReply-s: if vote granted by a quorum, returns collected replies.
//				Otherwise returns nil.
// error: ErrStaleLog, ErrStaleTermId, ErrTimeout.
// higherTerm: if seen, upgrade term and retry
func VoteOnce(
	candidate *LeaderId,
	logStatus logStat,
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

			lg.Infow("got res", "res.reply", res.reply, "res.err", res.err)

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
func (tr *TRaft) AddLog(cmd *Cmd) *Record {

	// TODO lock

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
			r.Overrides.Set(lsn)
			break
		}
	}

	if i == -1 {
		// there is not a interfering record.
		r.Overrides = NewTailBitmap(0)
	}

	// all log I do not know must be executed in order.
	// Because I do not know of the intefering relations.
	r.Depends = NewTailBitmap(tr.LogOffset)

	// reduce bitmap size by removing unknown logs
	r.Overrides.Union(NewTailBitmap(tr.LogOffset & ^63))

	tr.Logs = append(tr.Logs, r)

	return r
}

func (tr *TRaft) Vote(ctx context.Context, req *VoteReq) (*VoteReply, error) {
	repl := query(tr.actionCh, "vote", req).v.(*VoteReply)
	return repl, nil
}

// no lock protect, must be called by TRaft.Loop()
func (tr *TRaft) internalVote(req *VoteReq) *VoteReply {

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

	lg.Infow("handleVoteReq", "Id", id, "me.VotedFor", me.VotedFor, "req.Candidate", req.Candidate)

	if CmpLogStatus(req, me) < 0 {
		// I have more logs than the candidate.
		// It cant be a leader.
		tr.sendMsg("handle-vote-req:reject-by-logstat",
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
		tr.sendMsg("handle-vote-req:reject-by-term-id",
			"req.Candidate", req.Candidate,
			"me.VotedFor", me.VotedFor,
		)
		return repl
	}

	// grant vote
	me.VotedFor = req.Candidate.Clone()
	me.VoteExpireAt = uSecond() + leaderLease
	repl.VotedFor = req.Candidate.Clone()

	lg.Infow("voted", "id", id, "VotedFor", me.VotedFor)
	tr.sendMsg("handle-vote-req:grant",
		"req.Candidate", req.Candidate,
		"me.VotedFor", me.VotedFor)

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

func (tr *TRaft) Replicate(ctx context.Context, req *ReplicateReq) (*ReplicateReply, error) {

	// TODO use query() to get traft state

	// TODO persist stat change.
	// TODO lock

	fmt.Println("Replicate:", req)
	me := tr.Status[tr.Id]
	{
		r := me.VotedFor.Cmp(me.Committer)
		fmt.Println(me.VotedFor)
		fmt.Println(me.Committer)
		fmt.Println(r)
		if r < 0 {
			panic("wtf")
		}

	}

	repl := &ReplicateReply{
		VotedFor: me.VotedFor.Clone(),
	}

	// check leadership

	r := req.Committer.Cmp(me.VotedFor)
	if r < 0 {
		return repl, nil
	}

	if r > 0 {
		// r>0: it is a legal leader.
		// correctness will be kept.
		// follow the (literally) greatest leader!!! :DDD
		me.VotedFor = req.Committer.Clone()
	}

	// r == 0: Committer is the same as I voted for.

	if req.Committer.Cmp(me.Committer) > 0 {
		me.Committer = req.Committer.Clone()
	}

	// TODO: if a newer committer is seen, non-committed logs
	// can be sure to stale and should be cleaned.

	for _, r := range req.Logs {
		lsn := r.Seq
		for lsn-tr.LogOffset >= int64(len(tr.Logs)) {
			// fill in the gap
			tr.Logs = append(tr.Logs, &Record{})
		}
		tr.Logs[lsn-tr.LogOffset] = r

		// If a record interferes and overrides a previous log,
		// it then does not need the overrided log to be commit to commit this
		// record.
		// As if a previous log has already accepted.
		me.Accepted.Union(r.Overrides)
	}

	repl.Accepted = me.Accepted.Clone()
	repl.Committed = me.Committed.Clone()

	return repl, nil
}
