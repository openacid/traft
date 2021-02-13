package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLeaderId_Cmp(t *testing.T) {

	ta := require.New(t)

	cases := []struct {
		a, b *LeaderId
		want int
	}{
		{a: NewLeaderId(1, 1), b: NewLeaderId(1, 1), want: 0},
		{a: NewLeaderId(1, 2), b: NewLeaderId(1, 1), want: 1},
		{a: NewLeaderId(1, 0), b: NewLeaderId(1, 1), want: -1},
		{a: NewLeaderId(2, 0), b: NewLeaderId(1, 1), want: 1},
		{a: NewLeaderId(0, 0), b: NewLeaderId(1, 1), want: -1},

		{a: NewLeaderId(0, 0), b: nil, want: 0},
		{a: nil, b: NewLeaderId(1, 1), want: -1},
		{a: nil, b: nil, want: 0},
	}

	for i, c := range cases {
		got := c.a.Cmp(c.b)
		ta.Equal(c.want, got, "%d-th: case: %+v", i+1, c)
	}
}

func TestLeaderId_Clone(t *testing.T) {

	ta := require.New(t)

	a := NewLeaderId(1, 2)
	b := a.Clone()
	a.Term = 3
	a.Id = 4
	ta.Equal(int64(1), b.Term)
	ta.Equal(int64(2), b.Id)
}
