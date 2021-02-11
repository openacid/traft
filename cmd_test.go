package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCmdI64(t *testing.T) {

	ta := require.New(t)

	ta.Equal(&Cmd{
		Op:    "foo",
		Key:   "key",
		Value: &Cmd_VI64{3},
	}, NewCmdI64("foo", "key", 3))

}
