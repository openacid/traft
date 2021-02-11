package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newStatusAcc(t *testing.T) {

	ta := require.New(t)

	got := newStatusAcc(3, 4, 5)

	ta.Equal(int64(3), got.AcceptedFrom.Term)
	ta.Equal(int64(4), got.AcceptedFrom.Id)

	ta.Equal(int64(6), got.Accepted.Len())
}

func TestReplicaStatus_CmpAccepted(t *testing.T) {

	ta := require.New(t)

	cases := []struct {
		a, b *ReplicaStatus
		want int
	}{
		{a: newStatusAcc(1, 1, 1), b: newStatusAcc(1, 1, 1), want: 0},
		{a: newStatusAcc(1, 1, 2), b: newStatusAcc(1, 1, 1), want: 1},
		{a: newStatusAcc(1, 2, 0), b: newStatusAcc(1, 1, 1), want: 1},
		{a: newStatusAcc(2, 0, 0), b: newStatusAcc(1, 1, 1), want: 1},
	}

	for i, c := range cases {
		ta.Equal(c.want, c.a.CmpAccepted(c.b), "%d-th: case: %+v", i+1, c)
		ta.Equal(-c.want, c.b.CmpAccepted(c.a), "%d-th: case: %+v", i+1, c)
	}
}
