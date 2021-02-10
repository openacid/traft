package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTailBitmap(t *testing.T) {

	ta := require.New(t)

	cases := []struct {
		input int64
		want  *TailBitmap
	}{
		{
			input: 0,
			want: &TailBitmap{
				Offset:   0,
				Words:    make([]uint64, 0, 1024),
				Reclamed: 0,
			},
		},
		{
			input: 999999,
			want: &TailBitmap{
				Offset:   999999,
				Words:    make([]uint64, 0, 1024),
				Reclamed: 999999,
			},
		},
	}

	for i, c := range cases {
		got := NewTailBitmap(c.input)
		ta.Equal(c.want, got, "%d-th: case: %+v", i+1, c)
	}
}

func TestTailBitmap_Compact(t *testing.T) {

	ta := require.New(t)

	allOnes1024 := make([]uint64, 1024)
	for i, _ := range allOnes1024 {
		allOnes1024[i] = 0xffffffffffffffff
	}

	cases := []struct {
		input *TailBitmap
		want  *TailBitmap
	}{
		{
			input: &TailBitmap{
				Offset:   0,
				Words:    []uint64{0xffffffffffffffff},
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64,
				Words:    []uint64{},
				Reclamed: 0,
			},
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{0xffffffffffffffff},
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{},
				Reclamed: 0,
			},
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{0xffffffffffffffff, 1},
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1},
				Reclamed: 0,
			},
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    allOnes1024,
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64 * 1025,
				Words:    []uint64{},
				Reclamed: 64 * 1025,
			},
		},
	}

	for i, c := range cases {
		c.input.Compact()
		ta.Equal(c.want, c.input, "%d-th: case: %+v", i+1, c)
	}
}

func TestTailBitmap_Set(t *testing.T) {

	ta := require.New(t)

	allOnes1024 := make([]uint64, 1024)
	for i, _ := range allOnes1024 {
		allOnes1024[i] = 0xffffffffffffffff
	}

	cases := []struct {
		input *TailBitmap
		set   int64
		want  *TailBitmap
	}{
		{
			input: &TailBitmap{
				Offset:   0,
				Words:    []uint64{},
				Reclamed: 0,
			},
			set: 0,
			want: &TailBitmap{
				Offset:   0,
				Words:    []uint64{1},
				Reclamed: 0,
			},
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{},
				Reclamed: 0,
			},
			set: 65,
			want: &TailBitmap{
				Offset:   64,
				Words:    []uint64{2},
				Reclamed: 0,
			},
		},
		{
			input: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1},
				Reclamed: 0,
			},
			set: 5,
			want: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1},
				Reclamed: 0,
			},
		},
		{
			input: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1},
				Reclamed: 0,
			},
			set: 64*2 + 1,
			want: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{3},
				Reclamed: 0,
			},
		},
		{
			input: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1},
				Reclamed: 0,
			},
			set: 64*3 + 2,
			want: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1, 4},
				Reclamed: 0,
			},
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{0xffffffffffffff7f, 1},
				Reclamed: 0,
			},
			set: 64 + 7,
			want: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1},
				Reclamed: 0,
			},
		},
		{
			input: &TailBitmap{
				Offset:   64 * 1023,
				Words:    []uint64{0xffffffffffffff7f, 1},
				Reclamed: 0,
			},
			set: 64*1023 + 7,
			want: &TailBitmap{
				Offset:   64 * 1024,
				Words:    []uint64{1},
				Reclamed: 64 * 1024,
			},
		},
	}

	for i, c := range cases {
		c.input.Set(c.set)
		ta.Equal(c.want, c.input, "%d-th: case: %+v", i+1, c)
	}
}

func TestTailBitmap_Get(t *testing.T) {

	ta := require.New(t)

	allOnes1024 := make([]uint64, 1024)
	for i, _ := range allOnes1024 {
		allOnes1024[i] = 0xffffffffffffffff
	}

	cases := []struct {
		input *TailBitmap
		get   int64
		want  uint64
	}{
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{},
				Reclamed: 0,
			},
			get:  0,
			want: 1,
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{},
				Reclamed: 0,
			},
			get:  1,
			want: 2,
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{},
				Reclamed: 0,
			},
			get:  63,
			want: 1 << 63,
		},

		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{0xffffffffffffff7f, 1},
				Reclamed: 0,
			},
			get:  64 + 7,
			want: 0,
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{0xffffffffffffff7f, 1},
				Reclamed: 0,
			},
			get:  64 + 6,
			want: 1 << 6,
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{0xffffffffffffff7f, 1},
				Reclamed: 0,
			},
			get:  64 + 8,
			want: 1 << 8,
		},
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{0xffffffffffffff7f, 1},
				Reclamed: 0,
			},
			get:  64*2 + 0,
			want: 1,
		},
	}

	for i, c := range cases {
		got := c.input.Get(c.get)
		ta.Equal(c.want, got, "%d-th: Get case: %+v", i+1, c)

		got1 := c.input.Get1(c.get)
		if c.want != 0 {
			ta.Equal(uint64(1), got1, "%d-th: Get1 case: %+v", i+1, c)
		} else {
			ta.Equal(uint64(0), got1, "%d-th: Get1 case: %+v", i+1, c)
		}
	}
}

func TestTailBitmap_Clone(t *testing.T) {

	ta := require.New(t)

	allOnes1024 := make([]uint64, 1024)
	for i, _ := range allOnes1024 {
		allOnes1024[i] = 0xffffffffffffffff
	}

	cases := []struct {
		input *TailBitmap
	}{
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{1, 2, 3},
				Reclamed: 0,
			},
		},
	}

	for i, c := range cases {
		got := c.input.Clone()
		ta.Equal(c.input, got, "%d-th: same as cloned case: %+v", i+1, c)

		prev := c.input.Words[0]
		ta.NotEqual(1000, prev, "%d-th: not 1000 case: %+v", i+1, c)
		c.input.Words[0] = 1000
		ta.Equal(prev, got.Words[0], "%d-th: cloned does not change the original case: %+v", i+1, c)
	}
}

func TestTailBitmap_Union(t *testing.T) {

	ta := require.New(t)

	cases := []struct {
		input *TailBitmap
		other *TailBitmap
		want  *TailBitmap
	}{
		{
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{1},
				Reclamed: 0,
			},
			other: &TailBitmap{
				Offset:   64,
				Words:    []uint64{2},
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64,
				Words:    []uint64{3},
				Reclamed: 0,
			},
		},
		{
			// other is covered by input.Offset
			input: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1},
				Reclamed: 0,
			},
			other: &TailBitmap{
				Offset:   64,
				Words:    []uint64{2},
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1},
				Reclamed: 0,
			},
		},
		{
			// input is covered by other.Offset
			input: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1},
				Reclamed: 0,
			},
			other: &TailBitmap{
				Offset:   64 * 3,
				Words:    []uint64{2},
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64 * 3,
				Words:    []uint64{2},
				Reclamed: 64 * 3,
			},
		},
		{
			// 1111 xxxx xxxx xxxx
			// 1111 1111 yyyy
			input: &TailBitmap{
				Offset:   64,
				Words:    []uint64{1, 3, 7},
				Reclamed: 0,
			},
			other: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{8},
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{11, 7},
				Reclamed: 0,
			},
		},
		{
			// 1111 1111 xxxx xxxx xxxx
			// 1111 yyyy yyyy yyyy
			input: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{1, 3, 7},
				Reclamed: 0,
			},
			other: &TailBitmap{
				Offset:   64 * 1,
				Words:    []uint64{8, 2, 4},
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64 * 2,
				Words:    []uint64{3, 7, 7},
				Reclamed: 0,
			},
		},

		{
			// trigger reclaim if new all-ones are found.
			input: &TailBitmap{
				Offset:   64 * 1023,
				Words:    []uint64{1, 3, 7},
				Reclamed: 0,
			},
			other: &TailBitmap{
				Offset:   64 * 1023,
				Words:    []uint64{0xffffffffffffffff - 1},
				Reclamed: 0,
			},
			want: &TailBitmap{
				Offset:   64 * 1024,
				Words:    []uint64{3, 7},
				Reclamed: 64 * 1024,
			},
		},
	}

	for i, c := range cases {
		c.input.Union(c.other)
		ta.Equal(c.want, c.input, "%d-th: Get case: %+v", i+1, c)

	}
}
