package traft

// TRaftServer impl

import (
	context "context"
	sync "sync"

	"google.golang.org/protobuf/proto"
)

type TRaft struct {
	mu sync.Mutex
	Node
}

func (tr *TRaft) Vote(ctx context.Context, req *VoteReq) (*VoteReply, error) {

	candidate := req.Status
	me := tr.Status[tr.Id]

	// A vote reply just send back a voter's status.
	// It is the candidate's responsibility to check if a voter granted it.
	repl := &VoteReply{
		VoterStatus: &ReplicaStatus{
			VotedFor:     proto.Clone(me.VotedFor).(*LeaderId),
			AcceptedFrom: proto.Clone(me.AcceptedFrom).(*LeaderId),
			Accepted:     proto.Clone(me.Accepted).(*TailBitmap),
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
	me.VotedFor = proto.Clone(candidate.VotedFor).(*LeaderId)
	repl.VoterStatus.VotedFor = proto.Clone(candidate.VotedFor).(*LeaderId)

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
