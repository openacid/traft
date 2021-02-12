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

func TestCmd_Interfering(t *testing.T) {

	ta := require.New(t)

	cases := []struct {
		a, b *Cmd
		want bool
	}{
		{nil, nil, false},
		{nil, NewCmdI64("bar", "x", 4), false},
		{NewCmdI64("foo", "x", 3), nil, false},
		{NewCmdI64("foo", "x", 3), NewCmdI64("bar", "x", 4), false},
		{NewCmdI64("foo", "x", 3), NewCmdI64("foo", "x", 4), false},
		{NewCmdI64("set", "x", 3), NewCmdI64("set", "y", 4), false},
		{NewCmdI64("set", "x", 3), NewCmdI64("set", "x", 4), true},
	}

	for i, c := range cases {
		got := c.a.Interfering(c.b)
		ta.Equal(c.want, got, "%d-th: case: %+v", i+1, c)
	}
}
