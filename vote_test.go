package traft

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
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

// a helper func to setup TRaft cluster and close it.
// Because `defer tr.Stop()` does not block until the next test case
func withCluster(t *testing.T,
	name string,
	ids []int64,
	f func(t *testing.T, ts []*TRaft)) {

	lid := NewLeaderId

	ts := serveCluster(ids)
	for i, id := range ids {
		ts[i].initTraft(lid(0, 0), lid(0, 0), []int64{}, nil, nil, lid(0, id))
	}

	t.Run(name, func(t *testing.T) {
		f(t, ts)
	})

	stopAll(ts)
}

func TestTRaft_Vote(t *testing.T) {

	ta := require.New(t)

	bm := NewTailBitmap

	ids := []int64{1, 2, 3}
	id := int64(1)

	ts := serveCluster(ids)
	defer stopAll(ts)

	t1 := ts[0]

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
			Accepted:  bm(0, cand.logs...),
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
				allLogBitmap: bm(0, 5, 6),
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
				allLogBitmap: bm(0, 5, 6, 7),
				logs:         "[<>, <001#001:007{set(x, 7)}-0→0>]",
			},
		},

		// candidate has no upto date logs
		{
			candStat{candidateId: lid(2, 2), committer: lid(0, id), logs: []int64{5, 6}},
			voterStat{votedFor: lid(1, id), committer: lid(1, id), author: lid(1, id), logs: []int64{5, 6}},
			wantVoteReply{
				votedFor:     lid(1, id),
				committer:    lid(1, id),
				allLogBitmap: bm(0, 5, 6),
				logs:         "[]",
			},
		},

		// candidate has not enough logs
		// No log is sent back to candidate because it does not need to rebuild
		// full log history.
		{
			candStat{candidateId: lid(2, 2), committer: lid(1, id), logs: []int64{5}},
			voterStat{votedFor: lid(1, id), committer: lid(1, id), author: lid(1, id), logs: []int64{5, 6}},
			wantVoteReply{
				votedFor:     lid(1, id),
				committer:    lid(1, id),
				allLogBitmap: bm(0, 5, 6),
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
				allLogBitmap: bm(0, 5, 6),
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
				allLogBitmap: bm(0, 5, 6),
				logs:         "[]",
			},
		},
	}

	for i, c := range cases {
		fmt.Println("case-", i)
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

		ta.InDelta(uSecondI64()+leaderLease, t1.Status[id].VoteExpireAt, 1000*1000*1000)
	}
}

func TestTRaft_VoteOnce(t *testing.T) {

	// cluster = {0, 1, 2}
	// ts[0] vote once with differnt Committer/VotedFor settings.

	lid := NewLeaderId

	type wt struct {
		hasVoteReplies bool
		err            error
		higherTerm     int64
	}

	cases := []struct {
		name       string
		committers []*LeaderId
		votedFors  []*LeaderId
		logs       [][]string
		candidate  *LeaderId
		want       wt
	}{
		{name: "2emptyVoter/term-0",
			candidate: lid(0, 0),
			want: wt{
				hasVoteReplies: false,
				err:            ErrStaleTermId,
				higherTerm:     0,
			},
		},
		{name: "2emptyVoter/term-1",
			candidate: lid(1, 0),
			want: wt{
				hasVoteReplies: true,
				err:            nil,
				higherTerm:     -1,
			},
		},
		{name: "reject-by-one/stalelog",
			committers: []*LeaderId{nil, lid(2, 0)},
			votedFors:  []*LeaderId{nil, lid(2, 1)},
			candidate:  lid(1, 0),
			want: wt{
				hasVoteReplies: true,
				err:            nil,
				higherTerm:     -1,
			},
		},
		{name: "reject-by-one/higherTerm",
			committers: []*LeaderId{nil, nil, lid(0, 0)},
			votedFors:  []*LeaderId{nil, nil, lid(5, 2)},
			candidate:  lid(1, 0),
			want: wt{
				hasVoteReplies: true,
				err:            nil,
				higherTerm:     -1,
			},
		},
		{name: "reject-by-two/stalelog",
			committers: []*LeaderId{nil, lid(2, 0), lid(0, 0)},
			votedFors:  []*LeaderId{nil, lid(2, 1), lid(2, 2)},
			candidate:  lid(1, 0),
			want: wt{
				hasVoteReplies: false,
				err:            ErrStaleLog,
				higherTerm:     2,
			},
		},
		{name: "reject-by-two/stalelog-higherTerm",
			committers: []*LeaderId{nil, lid(2, 0), lid(0, 0)},
			votedFors:  []*LeaderId{nil, lid(2, 1), lid(5, 2)},
			logs:       [][]string{nil, nil, []string{"x=0"}},
			candidate:  lid(1, 0),
			want: wt{
				hasVoteReplies: false,
				err:            ErrStaleLog,
				higherTerm:     5,
			},
		},
		{name: "reject-by-two/higherTerm",
			votedFors: []*LeaderId{nil, lid(3, 1), lid(5, 2)},
			candidate: lid(1, 0),
			want: wt{
				hasVoteReplies: false,
				err:            ErrStaleTermId,
				higherTerm:     5,
			},
		},
	}

	for _, c := range cases {
		withCluster(t, c.name,
			[]int64{0, 1, 2},
			func(t *testing.T, ts []*TRaft) {
				ta := require.New(t)
				for i, cmt := range c.committers {
					if cmt != nil {
						ts[i].Status[int64(i)].Committer = cmt
					}
				}

				for i, v := range c.votedFors {
					if v != nil {
						ts[i].Status[int64(i)].VotedFor = v
					}
				}

				for i, ls := range c.logs {
					for _, l := range ls {
						ts[i].addlogs(l)
					}
				}

				voted, err, higher := VoteOnce(
					c.candidate,
					ExportLogStatus(ts[0].Status[0]),
					ts[0].Config.Clone(),
				)

				if c.want.hasVoteReplies {
					ta.NotNil(voted)
				} else {
					ta.Nil(voted)
				}
				ta.Equal(c.want.err, errors.Cause(err))
				ta.Equal(c.want.higherTerm, higher)
			})
	}
}

func TestTRaft_query(t *testing.T) {

	ta := require.New(t)

	ids := []int64{1}
	id1 := int64(1)
	lid := NewLeaderId

	ts := serveCluster(ids)
	defer stopAll(ts)

	t1 := ts[0]
	t1.initTraft(lid(1, 2), lid(3, 4), []int64{5}, nil, nil, lid(2, id1))

	got := t1.query(func() interface{} {
		return ExportLogStatus(t1.Status[t1.Id])
	}).v.(*LogStatus)
	ta.Equal("001#002", got.Committer.ShortStr())
	ta.Equal("0:20", got.Accepted.ShortStr())
}

func stopAll(ts []*TRaft) {
	for _, s := range ts {
		s.Stop()
	}
}

func readMsg(ts []*TRaft) string {

	// var msg string
	// select {
	// case msg = <-ts[0].MsgCh:
	// case msg = <-ts[1].MsgCh:
	// case msg = <-ts[2].MsgCh:
	// case <-time.After(time.Second):
	//     panic("timeout")
	// }

	// n TRaft and a timeout
	cases := make([]reflect.SelectCase, len(ts)+1)
	for i, t := range ts {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(t.MsgCh)}
	}
	cases[len(ts)] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(time.After(time.Second))}

	chosen, value, ok := reflect.Select(cases)
	// ok will be true if the channel has not been closed.
	if chosen == len(ts) {
		panic("timeout")
	}

	_ = ok

	msg := value.String()
	return msg
}

// waiting for expected message substring to present n times.
func waitForMsg(ts []*TRaft, msgs map[string]int) {
	for {
		msg := readMsg(ts)
		for s, _ := range msgs {
			if strings.Contains(msg, s) {
				msgs[s]--
				lg.Infow("got-msg", "msg", msg)
			}
		}

		all0 := true
		for _, n := range msgs {
			all0 = all0 && n == 0
		}

		lg.Infow("require-msg", "msgs", msgs)

		if all0 {
			return
		}
	}
}

func TestTRaft_VoteLoop(t *testing.T) {

	lid := NewLeaderId
	bm := NewTailBitmap

	withCluster(t, "emptyVoters/candidate-1",
		[]int64{0, 1, 2},
		func(t *testing.T, ts []*TRaft) {
			ta := require.New(t)

			go ts[0].VoteLoop()

			waitForMsg(ts, map[string]int{
				"vote-win 001#000": 1,
			})

			ta.Equal(lid(1, 0), ts[0].Status[0].VotedFor)
			ta.InDelta(uSecondI64()+leaderLease,
				ts[0].Status[0].VoteExpireAt, 1000*1000*1000)

			ta.Equal(lid(1, 0), ts[1].Status[1].VotedFor)
			ta.InDelta(uSecondI64()+leaderLease,
				ts[1].Status[1].VoteExpireAt, 1000*1000*1000)
		})

	withCluster(t, "emptyVoters/candidate-2",
		[]int64{0, 1, 2},
		func(t *testing.T, ts []*TRaft) {
			ta := require.New(t)

			go ts[1].VoteLoop()
			waitForMsg(ts, map[string]int{
				"vote-win 001#001": 1,
			})

			ta.Equal(lid(1, 1), ts[1].Status[1].VotedFor)

			ta.InDelta(uSecondI64()+leaderLease,
				ts[1].Status[1].VoteExpireAt, 1000*1000*1000)
		})

	withCluster(t, "emptyVoters/candidate-12",
		[]int64{0, 1, 2},
		func(t *testing.T, ts []*TRaft) {

			go ts[0].VoteLoop()
			go ts[1].VoteLoop()

			// only one succ to elect.
			// In 1 second, there wont be another winning election.
			waitForMsg(ts, map[string]int{
				"vote-win 001#001": 1,
			})
		})

	withCluster(t, "emptyVoters/candidate-123",
		[]int64{0, 1, 2},
		func(t *testing.T, ts []*TRaft) {

			go ts[0].VoteLoop()
			go ts[1].VoteLoop()
			go ts[2].VoteLoop()

			// only one succ to elect.
			// In 1 second, there wont be another winning election.
			waitForMsg(ts, map[string]int{
				"vote-win":  1,
				"vote-fail": 2,
			})
		})

	withCluster(t, "id2MaxCommitter",
		[]int64{0, 1, 2},
		func(t *testing.T, ts []*TRaft) {
			ts[0].initTraft(lid(2, 1), lid(0, 1), []int64{2}, nil, nil, lid(4, 0))
			ts[1].initTraft(lid(3, 2), lid(0, 1), []int64{2}, nil, nil, lid(4, 1))
			ts[2].initTraft(lid(1, 3), lid(0, 1), []int64{2}, nil, nil, lid(4, 2))

			go ts[0].VoteLoop()
			go ts[1].VoteLoop()
			go ts[2].VoteLoop()

			// only one succ to elect.
			// In 1 second, there wont be another winning election.
			waitForMsg(ts, map[string]int{
				"vote-win 005#001": 1,
				"vote-fail":        2,
			})
		})

	withCluster(t, "id2MaxLog",
		[]int64{0, 1, 2, 3, 4},
		func(t *testing.T, ts []*TRaft) {
			// we need 5 replica to collect different log from 2 replica
			ta := require.New(t)
			_ = ta

			// R0 0.2      Committer: 2-0
			// R1 0...4    Committer: 3-1
			// R2 n..3     Committer: 1-2
			ts[0].initTraft(lid(2, 0), lid(1, 1), []int64{0, 2}, nil, nil, lid(4, 0))
			ts[1].initTraft(lid(3, 1), lid(1, 1), []int64{0, 4}, nil, nil, lid(4, 1))
			ts[2].initTraft(lid(1, 2), lid(2, 1), []int64{0, 3}, nil, []int64{0}, lid(4, 2))
			// ts[3].initTraft(lid(1, 2), lid(1, 1), []int64{0, 2, 3}, nil, nil, lid(0, 3))
			// ts[4].initTraft(lid(1, 2), lid(1, 1), []int64{0, 2, 3}, nil, nil, lid(0, 4))

			ts[3].Stop()
			ts[4].Stop()
			ts[1].Status[1].VotedFor = lid(3, 1)
			go ts[1].VoteLoop()

			// only one succ to elect.
			// In 1 second, there wont be another winning election.
			waitForMsg(ts, map[string]int{
				"vote-win 005#001": 1,
			})

			ta.Equal(
				join("[<001#001:000{set(x, 0)}-0→0>",
					"<>",
					"<>",
					"<>",
					"<001#001:004{set(x, 4)}-0→0>]"),
				RecordsShortStr(ts[1].Logs, ""),
			)

			ta.Equal(lid(5, 1), ts[1].Status[1].Committer)
			ta.Equal(bm(0, 0, 4), ts[1].Status[1].Accepted)
			ta.Equal(bm(0), ts[1].Status[1].Committed)

			ta.Equal(lid(2, 0), ts[1].Status[0].Committer)
			// using Equal to avoid comparison between nil and []int64{}
			ta.True(bm(0).Equal(ts[1].Status[0].Accepted))
			ta.True(bm(0).Equal(ts[1].Status[0].Committed))

			ta.Equal(lid(1, 2), ts[1].Status[2].Committer)
			// reduced Accepted to Committed
			ta.Equal(bm(0, 0), ts[1].Status[2].Accepted)
			ta.Equal(bm(0, 0), ts[1].Status[2].Committed)
		})
}

func TestTRaft_Propose(t *testing.T) {

	lid := NewLeaderId
	bm := NewTailBitmap

	sendPropose := func(addr string, xcmd interface{}) *ProposeReply {
		cmd := toCmd(xcmd)
		var reply *ProposeReply
		rpcTo(addr, func(cli TRaftClient, ctx context.Context) {
			var err error
			reply, err = cli.Propose(ctx, cmd)
			if err != nil {
				lg.Infow("err:", "err", err)
			}
		})
		return reply
	}

	withCluster(t, "invalidLeader",
		[]int64{0, 1, 2},
		func(t *testing.T, ts []*TRaft) {
			ta := require.New(t)

			ts[0].initTraft(lid(2, 0), lid(1, 1), []int64{}, nil, nil, lid(2, 0))
			ts[1].initTraft(lid(3, 1), lid(1, 1), []int64{}, nil, nil, lid(3, 1))
			ts[2].initTraft(lid(1, 2), lid(2, 1), []int64{}, nil, []int64{0}, lid(1, 2))

			mems := ts[1].Config.Members

			// no leader elected, not allow to propose
			reply := sendPropose(mems[1].Addr, NewCmdI64("foo", "x", 1))
			ta.Equal(&ProposeReply{
				OK:          false,
				Err:         "vote expired",
				OtherLeader: nil,
			}, reply)

			// elect ts[1]
			go ts[1].VoteLoop()

			waitForMsg(ts, map[string]int{
				"vote-win 004#001": 1,
			})

			// send to non-leader replica:
			reply = sendPropose(mems[0].Addr, NewCmdI64("foo", "x", 1))
			ta.Equal(&ProposeReply{
				OK:          false,
				Err:         "I am not leader",
				OtherLeader: lid(4, 1)}, reply)
		})

	withCluster(t, "succ",
		[]int64{0, 1, 2},
		func(t *testing.T, ts []*TRaft) {

			ta := require.New(t)

			ts[0].initTraft(lid(2, 0), lid(1, 1), []int64{}, nil, nil, lid(3, 0))
			ts[1].initTraft(lid(3, 1), lid(1, 1), []int64{}, nil, nil, lid(3, 1))
			ts[2].initTraft(lid(1, 2), lid(2, 1), []int64{}, nil, []int64{0}, lid(3, 2))

			mems := ts[1].Config.Members

			// elect ts[1]
			go ts[1].VoteLoop()

			waitForMsg(ts, map[string]int{
				"vote-win 004#001": 1,
			})

			// TODO check state of other replicas

			// succ to propsoe
			reply := sendPropose(mems[1].Addr, "y=1")
			ta.Equal(&ProposeReply{OK: true}, reply)

			ta.Equal(bm(1), ts[1].Status[1].Accepted)
			ta.Equal(bm(1), ts[1].Status[1].Committed)
			ta.Equal(
				join("[<004#001:000{set(y, 1)}-0:1→0>", "]"),
				RecordsShortStr(ts[1].Logs, ""),
			)

			reply = sendPropose(mems[1].Addr, "y=2")
			ta.Equal(&ProposeReply{OK: true, OtherLeader: nil}, reply)

			ta.Equal(bm(2), ts[1].Status[1].Accepted)
			ta.Equal(bm(2), ts[1].Status[1].Committed)
			ta.Equal(
				join("[<004#001:000{set(y, 1)}-0:1→0>",
					"<004#001:001{set(y, 2)}-0:3→0>",
					"]"),
				RecordsShortStr(ts[1].Logs, ""),
			)

			reply = sendPropose(mems[1].Addr, "x=3")
			ta.Equal(&ProposeReply{OK: true, OtherLeader: nil}, reply)

			ta.Equal(bm(3), ts[1].Status[1].Accepted)
			ta.Equal(
				join("[<004#001:000{set(y, 1)}-0:1→0>",
					"<004#001:001{set(y, 2)}-0:3→0>",
					"<004#001:002{set(x, 3)}-0:4→0>",
					"]"),
				RecordsShortStr(ts[1].Logs, ""),
			)
		})
}

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

	ts[0].addlogs()
	ts[1].addlogs("x=0", "y=1", "x=2")
	ts[2].addlogs("", "y=5")

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

		wantOK        bool
		wantVotedFor  *LeaderId
		wantAccepted  *TailBitmap
		wantCommitted *TailBitmap
		wantLogs      []string
	}{
		{"unmatchedCommitter",
			0, lid(3, 0), sec1k, lid(1, 2), logs[0:],
			false, lid(3, 0), nil, nil, nil,
		},
		{"VotedForExpired",
			0, lid(2, 2), -sec1k, lid(1, 2), logs[0:],
			false, lid(2, 2), nil, nil, nil,
		},
		{"accept/log2",
			0, lid(3, 1), sec1k, lid(3, 1), logs[2:],
			true, lid(3, 1), bm(0, 0, 2), bm(0),
			[]string{
				"<>",
				"<>",
				"<005#001:002{set(x, 2)}-0:5→0>",
			},
		},
		{"accept/log12",
			0, lid(3, 1), sec1k, lid(3, 1), logs[1:],
			true, lid(3, 1), bm(3), bm(0),
			[]string{
				"<>",
				"<005#001:001{set(y, 1)}-0:2→0>",
				"<005#001:002{set(x, 2)}-0:5→0>",
			},
		},
		{"accept/log12/overrideOld",
			2, lid(3, 1), sec1k, lid(3, 1), logs[1:],
			true, lid(3, 1), bm(3), bm(1),
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
				})

				ta.Equal(c.wantOK, repl.OK)
				ta.Equal(c.wantVotedFor, repl.VotedFor)
				if c.wantAccepted != nil {
					ta.True(c.wantAccepted.Equal(repl.Accepted))
					ta.True(c.wantAccepted.Equal(dst.Accepted))
				}
				if c.wantCommitted != nil {
					ta.True(c.wantCommitted.Equal(repl.Committed))
					ta.True(c.wantCommitted.Equal(dst.Committed))
				}
				if c.wantLogs != nil {
					ta.Equal("["+join(c.wantLogs...)+"]",
						RecordsShortStr(ts[c.to].Logs, ""))
				}
			})
	}
}

func TestTRaft_AddLog_nil(t *testing.T) {

	ta := require.New(t)

	id := int64(1)
	tr := NewTRaft(id, map[int64]string{id: "123"})

	tr.addlogs("x=1", "y=1", nil, "x=1")

	ta.Equal(join(
		"[<000#001:000{set(x, 1)}-0:1→0>",
		"<000#001:001{set(y, 1)}-0:2→0>",
		"<>",
		"<000#001:003{set(x, 1)}-0:9→0>]"), RecordsShortStr(tr.Logs, ""))
}

func TestTRaft_AddLog(t *testing.T) {

	ta := require.New(t)

	id := int64(1)
	tr := NewTRaft(id, map[int64]string{id: "123"})

	tr.AddLog(NewCmdI64("set", "x", 1))
	ta.Equal("[<000#001:000{set(x, 1)}-0:1→0>]", RecordsShortStr(tr.Logs))

	tr.AddLog(NewCmdI64("set", "y", 1))
	ta.Equal(join(
		"[<000#001:000{set(x, 1)}-0:1→0>",
		"<000#001:001{set(y, 1)}-0:2→0>]"), RecordsShortStr(tr.Logs, ""))

	tr.AddLog(NewCmdI64("set", "x", 1))
	ta.Equal(join(
		"[<000#001:000{set(x, 1)}-0:1→0>",
		"<000#001:001{set(y, 1)}-0:2→0>",
		"<000#001:002{set(x, 1)}-0:5→0>]"), RecordsShortStr(tr.Logs, ""))

	varnames := "wxyz"

	for i := 0; i < 67; i++ {
		vi := i % len(varnames)
		tr.AddLog(NewCmdI64("set", varnames[vi:vi+1], int64(i)))
	}
	l := len(tr.Logs)
	ta.Equal("<000#001:069{set(y, 66)}-0:2222222222222222:22→0>", tr.Logs[l-1].ShortStr())

	// truncate some logs, then add another 67
	// To check Overrides and Depends

	tr.LogOffset = 65
	tr.Logs = tr.Logs[65:]

	for i := 0; i < 67; i++ {
		vi := i % len(varnames)
		tr.AddLog(NewCmdI64("set", varnames[vi:vi+1], 100+int64(i)))
	}
	l = len(tr.Logs)
	ta.Equal("<000#001:136{set(y, 166)}-64:1111111111111122:111→64:1>", tr.Logs[l-1].ShortStr())

}
