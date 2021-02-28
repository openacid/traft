package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVoteReply_Pop(t *testing.T) {

	ta := require.New(t)

	vr := &ElectReply{
		Logs: []*LogRecord{
			NewRecord(NewLeaderId(1, 2), 5, nil),
			NewRecord(NewLeaderId(1, 2), 7, nil),
		},
	}

	var got *LogRecord
	ta.Nil(vr.PopRecord(4))

	got = vr.PopRecord(5)
	ta.NotNil(got)
	ta.Equal(int64(5), got.Seq)

	// pop again
	ta.Nil(vr.PopRecord(5))

	ta.Nil(vr.PopRecord(6))

	got = vr.PopRecord(7)
	ta.NotNil(got)
	ta.Equal(int64(7), got.Seq)

	// pop from empty logs:
	ta.Nil(vr.PopRecord(5))
}
