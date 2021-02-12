package traft

// TRaftServer impl

import (
	context "context"
	fmt "fmt"
	sync "sync"
	// "google.golang.org/protobuf/proto"
)

type TRaft struct {
	// TODO lock first
	mu sync.Mutex
	Node
}

// init a TRaft for test, all logs are `set x=lsn`
func (tr *TRaft) initLog(
	// proposer of the logs
	committer *LeaderId,
	// author of the logs
	author *LeaderId,
	// log seq numbers to generate.
	accepted []int64,
	nilLogs map[int64]bool,
	committed []int64,
	votedFor *LeaderId,
) {
	id := tr.Id

	tr.LogOffset, tr.Log = initLog2(author, accepted, nilLogs)

	tr.Status[id].Committer = committer.Clone()
	tr.Status[id].Accepted = NewTailBitmap(0, accepted...)

	if committed == nil {
		tr.Status[id].Committed = NewTailBitmap(0)
	} else {
		tr.Status[id].Committed = NewTailBitmap(0, committed...)
	}

	tr.Status[id].VotedFor = votedFor.Clone()
}

func initLog2(
	// author of the logs
	author *LeaderId,
	// log seq numbers to generate.
	accepted []int64,
	nilLogs map[int64]bool,
) (int64, []*Record) {
	logs := make([]*Record, 0)
	if len(accepted) == 0 {
		return 0, logs
	}

	last := accepted[len(accepted)-1]
	start := accepted[0]
	for i := start; i <= last; i++ {
		logs = append(logs, &Record{})
	}

	for _, lsn := range accepted {
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

func (tr *TRaft) Vote(ctx context.Context, req *VoteReq) (*VoteReply, error) {

	me := tr.Status[tr.Id]

	// A vote reply just send back a voter's status.
	// It is the candidate's responsibility to check if a voter granted it.
	repl := &VoteReply{
		VotedFor:  me.VotedFor.Clone(),
		Committer: me.Committer.Clone(),
		Accepted:  me.Accepted.Clone(),
	}

	if CmpLogStatus(req, me) < 0 {
		// I have more logs than the candidate.
		// It cant be a leader.
		return repl, nil
	}

	// candidate has the upto date logs.

	r := req.Candidate.Cmp(me.VotedFor)
	if r < 0 {
		// I've voted for other leader with higher privilege.
		// This candidate could not be a legal leader.
		// just send back enssential info to info it.
		return repl, nil
	}

	// grant vote
	me.VotedFor = req.Candidate.Clone()
	repl.VotedFor = req.Candidate.Clone()

	// send back the logs I have but the candidate does not.

	logs := make([]*Record, 0)

	start := me.Accepted.Offset
	end := me.Accepted.Len()
	for i := start; i < end; i++ {
		if me.Accepted.Get(i) != 0 && req.Accepted.Get(i) == 0 {
			fmt.Println("append:", tr.Log[i-tr.LogOffset].ShortStr())
			logs = append(logs, tr.Log[i-tr.LogOffset])
		}
	}

	repl.Logs = logs

	return repl, nil
}

func (tr *TRaft) Replicate(ctx context.Context, req *ReplicateReq) (*ReplicateReply, error) {

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
		for lsn-tr.LogOffset >= int64(len(tr.Log)) {
			// fill in the gap
			tr.Log = append(tr.Log, &Record{})
		}
		tr.Log[lsn-tr.LogOffset] = r

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
