package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRecord(t *testing.T) {

	ta := require.New(t)

	ta.Equal(&LogRecord{
		Author: &LeaderId{1, 2},
		Seq:    3,
		Cmd: &Cmd{
			Op:    "foo",
			Key:   "key",
			Value: &Cmd_VI64{4},
		},
	}, NewRecord(NewLeaderId(1, 2), 3, NewCmdI64("foo", "key", 4)))

}

func TestRecord_Interfering(t *testing.T) {

	ta := require.New(t)

	lid := NewLeaderId
	cmd := NewCmdI64

	cases := []struct {
		a, b *LogRecord
		want bool
	}{
		{nil, nil, false},
		{nil, NewRecord(lid(0, 1), 0, cmd("bar", "x", 1)), false},
		{NewRecord(lid(0, 1), 0, cmd("foo", "x", 1)), NewRecord(lid(0, 1), 0, nil), false},
		{NewRecord(lid(0, 1), 0, cmd("foo", "x", 1)), NewRecord(lid(0, 1), 0, cmd("bar", "x", 1)), false},
		{NewRecord(lid(0, 1), 0, cmd("set", "x", 1)), NewRecord(lid(0, 1), 0, cmd("set", "y", 1)), false},
		{NewRecord(lid(0, 1), 0, cmd("set", "x", 1)), NewRecord(lid(0, 1), 0, cmd("set", "x", 1)), true},
	}

	for i, c := range cases {
		ta.Equal(c.want, c.a.Interfering(c.b), "%d-th: case: %+v", i+1, c)
		ta.Equal(c.want, c.b.Interfering(c.a), "%d-th: case: %+v", i+1, c)
	}
}
