package traft

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTRaft_Vote(t *testing.T) {

	ta := require.New(t)

	ids := []int64{1, 2, 3}

	servers, trafts := serveCluster(ids)
	id := int64(1)

	defer func() {
		for _, s := range servers {
			s.Stop()
		}
	}()

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
	}

	type wantStat struct {
		votedFor     *LeaderId
		committer    *LeaderId
		allLogBitmap *TailBitmap
		logs         string
	}

	testVote := func(
		cand candStat,
		voter voterStat,
		want wantStat,
	) *VoteReply {

		t1 := trafts[0]

		t1.initLog(
			voter.committer, voter.author, voter.logs,
			voter.votedFor,
		)

		req := &VoteReq{
			Candidate: cand.candidateId,
			Committer: cand.committer,
			Accepted:  NewTailBitmap(0, cand.logs...),
		}

		var reply *VoteReply
		addr := t1.Config.GetReplicaInfo(id).Addr

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
		want  wantStat
	}{
		// vote granted
		{
			candStat{candidateId: lid(2, 2), committer: lid(1, id), logs: []int64{5}},
			voterStat{votedFor: lid(0, id), committer: lid(0, id), author: lid(1, id), logs: []int64{5, 6}},
			wantStat{
				votedFor:     lid(2, 2),
				committer:    lid(0, id),
				allLogBitmap: NewTailBitmap(0, 5, 6),
				logs:         "[<001#001:006{set(x, 6)}>]",
			},
		},

		// candidate has no upto date logs
		{
			candStat{candidateId: lid(2, 2), committer: lid(0, id), logs: []int64{5, 6}},
			voterStat{votedFor: lid(0, id), committer: lid(1, id), author: lid(1, id), logs: []int64{5, 6}},
			wantStat{
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
			wantStat{
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
			wantStat{
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
			wantStat{
				votedFor:     lid(3, id),
				committer:    lid(1, id),
				allLogBitmap: NewTailBitmap(0, 5, 6),
				logs:         "[]",
			},
		},
	}

	for i, c := range cases {
		reply := testVote(c.cand, c.voter, c.want)

		fmt.Println(reply.String())
		fmt.Println(RecordsShortStr(reply.Logs))

		ta.Equal(
			c.want,
			wantStat{
				votedFor:     reply.VotedFor,
				committer:    reply.Committer,
				allLogBitmap: reply.Accepted,
				logs:         RecordsShortStr(reply.Logs),
			},
			"%d-th: case: %+v", i+1, c)
	}
}
