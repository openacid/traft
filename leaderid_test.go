package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLeaderId_Cmp(t *testing.T) {

	ta := require.New(t)

	cases := []struct {
		a, b LeaderId
		want int
	}{
		{a: LeaderId{Term: 1, Id: 1}, b: LeaderId{Term: 1, Id: 1}, want: 0},
		{a: LeaderId{Term: 1, Id: 2}, b: LeaderId{Term: 1, Id: 1}, want: 1},
		{a: LeaderId{Term: 1, Id: 0}, b: LeaderId{Term: 1, Id: 1}, want: -1},
		{a: LeaderId{Term: 2, Id: 0}, b: LeaderId{Term: 1, Id: 1}, want: 1},
		{a: LeaderId{Term: 0, Id: 0}, b: LeaderId{Term: 1, Id: 1}, want: -1},
	}

	for i, c := range cases {
		got := c.a.Cmp(&c.b)
		ta.Equal(c.want, got, "%d-th: case: %+v", i+1, c)
	}
}
