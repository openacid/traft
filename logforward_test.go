package traft

import (
	context "context"
	fmt "fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTRaft_LogForward(t *testing.T) {

	ta := require.New(t)
	_ = ta

	ids := []int64{0, 1, 2}

	ts := serveCluster(ids)
	defer stopAll(ts)

	lid := NewLeaderId
	bm := NewTailBitmap

	// init cluster
	// give ts[1] a highest term thus to be a leader
	ts[0].initTraft(lid(2, 0), lid(0, 1), []int64{}, nil, nil, lid(3, 0))
	ts[1].initTraft(lid(3, 1), lid(0, 1), []int64{}, nil, nil, lid(5, 1))
	ts[2].initTraft(lid(1, 2), lid(0, 1), []int64{}, nil, []int64{0}, lid(2, 2))

	ts[0].addLogs()
	ts[1].addLogs("x=0", "y=1", "x=2")
	ts[2].addLogs("", "y=5")

	sendLogForward := func(addr string, req *LogForwardReq) *LogForwardReply {
		var reply *LogForwardReply
		rpcTo(addr, func(cli TRaftClient, ctx context.Context) {
			var err error
			reply, err = cli.LogForward(ctx, req)
			if err != nil {
				lg.Infow("sendLogForward:err", "err", err)
			}
		})
		return reply
	}

	logs := ts[1].Logs

	sec1k := int64(time.Second * 1000)
	cases := []struct {
		name     string
		to       int64
		votedFor *LeaderId
		expire   int64

		committer *LeaderId
		logs      []*Record
		committed *TailBitmap

		wantOK        bool
		wantVotedFor  *LeaderId
		wantAccepted  *TailBitmap
		wantCommitted *TailBitmap
		wantLogs      []string
	}{
		{"unmatchedCommitter",
			0, lid(3, 0), sec1k,
			lid(1, 2), logs[0:], nil,
			false, lid(3, 0), nil, nil, nil,
		},
		{"VotedForExpired",
			0, lid(2, 2), -sec1k,
			lid(1, 2), logs[0:], nil,
			false, lid(2, 2), nil, nil, nil,
		},
		{"accept/log2",
			0, lid(3, 1), sec1k,
			lid(3, 1), logs[2:], nil,
			true, lid(3, 1), bm(0, 0, 2), bm(0),
			[]string{
				"<>",
				"<>",
				"<005#001:002{set(x, 2)}-0:5→0>",
			},
		},
		{"accept/log12",
			0, lid(3, 1), sec1k,
			lid(3, 1), logs[1:], nil,
			true, lid(3, 1), bm(3), bm(0),
			[]string{
				"<>",
				"<005#001:001{set(y, 1)}-0:2→0>",
				"<005#001:002{set(x, 2)}-0:5→0>",
			},
		},
		{"accept/log12/overrideOld",
			2, lid(3, 1), sec1k,
			lid(3, 1), logs[1:], nil,
			true, lid(3, 1), bm(3), bm(1),
			[]string{
				"<>",
				"<005#001:001{set(y, 1)}-0:2→0>",
				"<005#001:002{set(x, 2)}-0:5→0>",
			},
		},
		{"accept/log12/mergeCommitted",
			2, lid(3, 1), sec1k,
			lid(3, 1), logs[1:], bm(0, 2, 3),
			true, lid(3, 1), bm(3), bm(0, 0, 2),
			[]string{
				"<>",
				"<005#001:001{set(y, 1)}-0:2→0>",
				"<005#001:002{set(x, 2)}-0:5→0>",
			},
		},
	}

	for _, c := range cases {
		t.Run(
			fmt.Sprintf("%d-to-%d/%s", 1, c.to, c.name),
			func(t *testing.T) {
				dst := ts[c.to].Status[c.to]
				dst.VotedFor = c.votedFor
				dst.VoteExpireAt = uSecondI64() + c.expire

				fmt.Println(ts[c.to].Node)
				ts[c.to].checkStatus()

				addr := ts[1].Config.Members[c.to].Addr
				repl := sendLogForward(addr, &LogForwardReq{
					Committer: c.committer,
					Logs:      c.logs,
					Committed: c.committed,
				})

				ta.Equal(c.wantOK, repl.OK)

				ta.Equal(c.wantVotedFor, repl.VotedFor)
				if c.wantAccepted != nil {
					ta.Equal(c.wantAccepted, repl.Accepted.Normalize())
					ta.Equal(c.wantAccepted, dst.Accepted.Normalize())
				}

				if c.wantCommitted != nil {
					ta.Equal(c.wantCommitted, repl.Committed.Normalize())
					ta.Equal(c.wantCommitted, dst.Committed.Normalize())
				}

				if c.wantLogs != nil {
					ta.Equal("["+join(c.wantLogs...)+"]",
						RecordsShortStr(ts[c.to].Logs, ""))
				}
			})
	}
}
