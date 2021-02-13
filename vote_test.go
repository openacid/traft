package traft

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type candStat struct {
	candidateId *LeaderId
	committer   *LeaderId
	logs        []int64
}

type voterStat struct {
	votedFor  *LeaderId
	committer *LeaderId
	author    *LeaderId
	logs      []int64
	nilLogs   map[int64]bool
	committed []int64
}

type wantVoteReply struct {
	votedFor     *LeaderId
	committer    *LeaderId
	allLogBitmap *TailBitmap
	logs         string
}

type replicateReqStat struct {
	committer *LeaderId
	logs      []int64
	nilLogs   map[int64]bool
}

type wantReplicateReply struct {
	votedFor  *LeaderId
	accepted  *TailBitmap
	committed *TailBitmap
}

type wantVoterStat struct {
	votedFor *LeaderId
	accepted *TailBitmap
	logs     string
}

func TestTRaft_Vote(t *testing.T) {

	ta := require.New(t)

	ids := []int64{1, 2, 3}
	id := int64(1)

	servers, trafts := serveCluster(ids)
	defer func() {
		for _, s := range servers {
			s.Stop()
		}
	}()

	t1 := trafts[0]

	testVote := func(
		cand candStat,
		voter voterStat,
	) *VoteReply {

		t1.initTraft(
			voter.committer, voter.author, voter.logs, voter.nilLogs, nil,
			voter.votedFor,
		)

		req := &VoteReq{
			Candidate: cand.candidateId,
			Committer: cand.committer,
			Accepted:  NewTailBitmap(0, cand.logs...),
		}

		var reply *VoteReply
		addr := t1.Config.Members[id].Addr

		rpcTo(addr, func(cli TRaftClient, ctx context.Context) {
			var err error
			reply, err = cli.Vote(ctx, req)
			if err != nil {
				panic("wtf")
			}
		})

		return reply
	}

	lid := NewLeaderId

	cases := []struct {
		cand  candStat
		voter voterStat
		want  wantVoteReply
	}{
		// vote granted
		{
			candStat{candidateId: lid(2, 2), committer: lid(1, id), logs: []int64{5}},
			voterStat{votedFor: lid(0, id), committer: lid(0, id), author: lid(1, id), logs: []int64{5, 6}},
			wantVoteReply{
				votedFor:     lid(2, 2),
				committer:    lid(0, id),
				allLogBitmap: NewTailBitmap(0, 5, 6),
				logs:         "[<001#001:006{set(x, 6)}-0→0>]",
			},
		},

		// vote granted
		// send back nil logs
		{
			candStat{candidateId: lid(2, 2), committer: lid(1, id), logs: []int64{5}},
			voterStat{votedFor: lid(0, id), committer: lid(0, id), author: lid(1, id), logs: []int64{5, 6, 7}, nilLogs: map[int64]bool{6: true}},
			wantVoteReply{
				votedFor:     lid(2, 2),
				committer:    lid(0, id),
				allLogBitmap: NewTailBitmap(0, 5, 6, 7),
				logs:         "[<>, <001#001:007{set(x, 7)}-0→0>]",
			},
		},

		// candidate has no upto date logs
		{
			candStat{candidateId: lid(2, 2), committer: lid(0, id), logs: []int64{5, 6}},
			voterStat{votedFor: lid(0, id), committer: lid(1, id), author: lid(1, id), logs: []int64{5, 6}},
			wantVoteReply{
				votedFor:     lid(0, id),
				committer:    lid(1, id),
				allLogBitmap: NewTailBitmap(0, 5, 6),
				logs:         "[]",
			},
		},

		// candidate has not enough logs
		// No log is sent back to candidate because it does not need to rebuild
		// full log history.
		{
			candStat{candidateId: lid(2, 2), committer: lid(1, id), logs: []int64{5}},
			voterStat{votedFor: lid(0, id), committer: lid(1, id), author: lid(1, id), logs: []int64{5, 6}},
			wantVoteReply{
				votedFor:     lid(0, id),
				committer:    lid(1, id),
				allLogBitmap: NewTailBitmap(0, 5, 6),
				logs:         "[]",
			},
		},

		// candidate has smaller term.
		// No log sent back.
		{
			candStat{candidateId: lid(2, 2), committer: lid(1, id), logs: []int64{5, 6}},
			voterStat{votedFor: lid(3, id), committer: lid(1, id), author: lid(1, id), logs: []int64{5, 6}},
			wantVoteReply{
				votedFor:     lid(3, id),
				committer:    lid(1, id),
				allLogBitmap: NewTailBitmap(0, 5, 6),
				logs:         "[]",
			},
		},

		// candidate has smaller id.
		// No log sent back.
		{
			candStat{candidateId: lid(3, id-1), committer: lid(1, id), logs: []int64{5, 6}},
			voterStat{votedFor: lid(3, id), committer: lid(1, id), author: lid(1, id), logs: []int64{5, 6}},
			wantVoteReply{
				votedFor:     lid(3, id),
				committer:    lid(1, id),
				allLogBitmap: NewTailBitmap(0, 5, 6),
				logs:         "[]",
			},
		},
	}

	for i, c := range cases {
		reply := testVote(c.cand, c.voter)

		fmt.Println(reply.String())
		fmt.Println(RecordsShortStr(reply.Logs))

		ta.Equal(
			c.want,
			wantVoteReply{
				votedFor:     reply.VotedFor,
				committer:    reply.Committer,
				allLogBitmap: reply.Accepted,
				logs:         RecordsShortStr(reply.Logs),
			},
			"%d-th: case: %+v", i+1, c)
	}
}

func TestTRaft_Replicate(t *testing.T) {

	ta := require.New(t)

	ids := []int64{1, 2, 3}
	id := int64(1)

	servers, trafts := serveCluster(ids)
	defer func() {
		for _, s := range servers {
			s.Stop()
		}
	}()

	t1 := trafts[0]
	me := t1.Status[id]

	testReplicate := func(
		rreq replicateReqStat,
		voter voterStat,
	) *ReplicateReply {

		t1.initTraft(
			voter.committer, voter.author, voter.logs, voter.nilLogs, voter.committed,
			voter.votedFor,
		)

		_, logs := buildPseudoLogs(rreq.committer, rreq.logs, rreq.nilLogs)
		req := &ReplicateReq{
			Committer: rreq.committer,
			Logs:      logs,
		}

		var reply *ReplicateReply
		addr := t1.Config.Members[id].Addr

		rpcTo(addr, func(cli TRaftClient, ctx context.Context) {
			var err error
			reply, err = cli.Replicate(ctx, req)
			if err != nil {
				panic("wtf")
			}
		})

		return reply
	}

	lid := NewLeaderId
	_ = lid

	cases := []struct {
		cand      replicateReqStat
		voter     voterStat
		want      wantReplicateReply
		wantVoter wantVoterStat
	}{
		//
		// {
		//     replicateReqStat{committer: lid(1, id), logs: []int64{5}},
		//     voterStat{votedFor: lid(1, id), committer: lid(1, id), author: lid(1, id), logs: []int64{}},
		//     wantReplicateReply{
		//         votedFor:  lid(1, id),
		//         accepted:  NewTailBitmap(0, 5),
		//         committed: NewTailBitmap(0),
		//     },
		//     wantVoterStat{
		//         votedFor: lid(1, id),
		//         accepted: NewTailBitmap(0, 5),
		//         logs:     "",
		//     },
		// },
	}

	for i, c := range cases {
		reply := testReplicate(c.cand, c.voter)

		fmt.Println(reply.String())

		ta.Equal(
			c.want,
			wantReplicateReply{
				votedFor:  reply.VotedFor,
				accepted:  reply.Accepted,
				committed: reply.Committed,
			},
			"%d-th: reply: case: %+v", i+1, c)

		ta.Equal(
			c.want,
			wantVoterStat{
				votedFor: me.VotedFor,
				accepted: me.Accepted,
				logs:     RecordsShortStr(t1.Logs),
			},
			"%d-th: voter: case: %+v", i+1, c)
	}
}

func TestTRaft_AddLog(t *testing.T) {

	ta := require.New(t)

	id := int64(1)
	tr := &TRaft{Node: *NewNode(id, map[int64]string{id: "123"})}
	tr.AddLog(NewCmdI64("set", "x", 1))
	// me := tr.Status[id]

	ta.Equal("[<000#001:000{set(x, 1)}-0→0>]", RecordsShortStr(tr.Logs))

	varnames := "wxyz"

	for i := 0; i < 67; i++ {
		vi := i % len(varnames)
		tr.AddLog(NewCmdI64("set", varnames[vi:vi+1], int64(i)))
	}
	l := len(tr.Logs)
	ta.Equal("<000#001:067{set(y, 66)}-0:8888888888888880:8→0>", tr.Logs[l-1].ShortStr())

	// truncate some logs, then add another 67
	// To check Overrides and Depends

	tr.LogOffset = 65
	tr.Logs = tr.Logs[65:]

	for i := 0; i < 67; i++ {
		vi := i % len(varnames)
		tr.AddLog(NewCmdI64("set", varnames[vi:vi+1], 100+int64(i)))
	}
	l = len(tr.Logs)
	ta.Equal("<000#001:134{set(y, 166)}-64:4444444444444448:44→64:1>", tr.Logs[l-1].ShortStr())

}
