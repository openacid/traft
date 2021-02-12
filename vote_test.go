package traft

import (
	context "context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTRaft_Vote(t *testing.T) {

	ta := require.New(t)
	// TODO
	_ = ta

	ids := []int64{1, 2, 3}

	servers, trafts := serveCluster(ids)

	defer func() {
		for _, s := range servers {
			s.Stop()
		}
	}()

	{
		// the first node:
		//           13
		//
		//   6₀₃₋₅₁
		//   5₀₃₋₅₁  5₁₃₋??
		//   --- --- ---
		//   t1      c
		id := int64(1)
		t1 := trafts[0]

		candidate := NewLeaderId(1, 3)

		t1.initLog(
			NewLeaderId(0, 3),
			NewLeaderId(5, id),
			[]int64{5, 6},
			NewLeaderId(0, 2),
		)
		r6 := t1.Log[1]

		req := &VoteReq{
			Status: &ReplicaStatus{
				VotedFor: candidate,
				// the candidate has latest log from leader (1, 3)
				AcceptedFrom: NewLeaderId(1, 3),
				// candidate has log [0, 6)
				Accepted: NewTailBitmap(0, 5),
			},
		}

		addr := t1.Config.GetReplicaInfo(id).Addr
		rpcTo(addr, func(cli TRaftClient, ctx context.Context) {

			reply, err := cli.Vote(ctx, req)
			if err != nil {
				panic("wtf")
			}

			ta.Equal(
				&VoteReply{
					VoterStatus: &ReplicaStatus{
						VotedFor:     NewLeaderId(1, 3),
						AcceptedFrom: NewLeaderId(0, 3),
						Accepted:     NewTailBitmap(0, 5, 6),
					},
					Logs: []*Record{
						r6,
					},
				}, reply)

		})
	}

}
