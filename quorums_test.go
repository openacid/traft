package traft

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildMajorityQuorums(t *testing.T) {

	ta := require.New(t)

	cases := []struct {
		input uint64
		want  []string
	}{
		{
			input: 7,
			want: []string{
				"11000000",
				"10100000",
				"01100000",
			},
		},
		{
			input: 1 + 1<<2 + 1<<3,
			want: []string{
				"10100000",
				"10010000",
				"00110000",
			},
		},
		{
			input: 1 + 1<<3 + 1<<4,
			want: []string{
				"10010000",
				"10001000",
				"00011000",
			},
		},
		{
			input: 1<<2 + 1<<3 + 1<<4 + 1<<5,
			want: []string{
				"00111000",
				"00110100",
				"00101100",
				"00011100",
			},
		},
	}

	for i, c := range cases {
		got := buildMajorityQuorums(c.input)
		gotStr := fmtBitmap(got)
		ta.Equal(c.want, gotStr, "%d-th: case: %+v", i+1, c)
	}
}

func fmtBitmap(vs []uint64) []string {
	rst := make([]string, 0)
	for _, v := range vs {
		rst = append(rst, fmt.Sprintf("%08b", bits.Reverse8(byte(v))))
	}
	return rst
}
