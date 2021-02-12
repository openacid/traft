package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRecord(t *testing.T) {

	ta := require.New(t)

	ta.Equal(&Record{
		Author: &LeaderId{1, 2},
		Seq:    3,
		Cmd: &Cmd{
			Op:    "foo",
			Key:   "key",
			Value: &Cmd_VI64{4},
		},
	}, NewRecord(NewLeaderId(1, 2), 3, NewCmdI64("foo", "key", 4)))

}
