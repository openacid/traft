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
	}

	for i, c := range cases {
		got := c.a.Cmp(c.b)
		ta.Equal(c.want, got, "%d-th: case: %+v", i+1, c)
	}
}
