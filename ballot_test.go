package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBallot(t *testing.T) {

	ta := require.New(t)

	got := NewBallot(1, 2, 3, 4)
	ta.Equal(int64(1), got.Current.Term)
	ta.Equal(int64(2), got.Current.Id)
	ta.Equal(int64(3), got.Accepted.Term)
	ta.Equal(int64(4), got.Accepted.Id)
}
