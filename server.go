package traft

// TRaftServer impl

import (
	context "context"
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
	from *LeaderId,
	// creator of the logs
	creator *LeaderId,
	// log seq numbers to generate.
	accepted []int64,
	votedFor *LeaderId,
) {
	id := tr.Id

	tr.LogOffset = accepted[0]
	tr.Log = make([]*Record, 0)
	for _, lsn := range accepted {
		tr.Log = append(tr.Log,
			NewRecord(creator.Clone(),
				lsn,
				NewCmdI64("set", "x", lsn)))
	}

	tr.Status[id].AcceptedFrom = from.Clone()
	tr.Status[id].Accepted = NewTailBitmap(0, accepted...)

	tr.Status[id].VotedFor = votedFor.Clone()
}

func (tr *TRaft) Vote(ctx context.Context, req *VoteReq) (*VoteReply, error) {

	candidate := req.Status
	me := tr.Status[tr.Id]

	// A vote reply just send back a voter's status.
	// It is the candidate's responsibility to check if a voter granted it.
	repl := &VoteReply{
		VoterStatus: &ReplicaStatus{
			VotedFor:     me.VotedFor.Clone(),
			AcceptedFrom: me.AcceptedFrom.Clone(),
			Accepted:     me.Accepted.Clone(),
		},
	}

	if candidate.CmpAccepted(me) < 0 {
		// I have more logs than the candidate.
		// It cant be a leader.
		return repl, nil
	}

	// candidate has the upto date logs.

	r := candidate.VotedFor.Cmp(me.VotedFor)
	if r < 0 {
		// I've voted for other leader with higher privilege.
		// This candidate could not be a legal leader.
		// just send back enssential info to info it.
		return repl, nil
	}

	// grant vote
	me.VotedFor = candidate.VotedFor.Clone()
	repl.VoterStatus.VotedFor = candidate.VotedFor.Clone()

	// send back the logs I have but the candidate does not.

	logs := make([]*Record, 0)

	start := me.Accepted.Offset
	end := me.Accepted.Len()
	for i := start; i < end; i++ {
		if me.Accepted.Get(i) != 0 && candidate.Accepted.Get(i) == 0 {
			logs = append(logs, tr.Log[i-tr.LogOffset])
		}
	}

	repl.Logs = logs

	return repl, nil
}

func (tr *TRaft) Replicate(ctx context.Context, req *Record) (*ReplicateReply, error) {
	return nil, nil
}
