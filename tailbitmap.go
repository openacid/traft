package traft

import "github.com/openacid/low/bitmap"

// reclaimThreshold is the size threshold in bit for reclamation of `Words`.
var reclaimThreshold = int64(1024) * 64

// NewTailBitmap creates an TailBitmap with a preset Offset and an empty
// tail bitmap.
//
// Since 0.1.22
func NewTailBitmap(offset int64) *TailBitmap {
	tb := &TailBitmap{
		Offset:   offset,
		Reclamed: offset,
		Words:    make([]uint64, 0, reclaimThreshold>>6),
	}
	return tb
}

// Compact all leading all-ones words in the bitmap.
//
// Since 0.1.22
func (tb *TailBitmap) Compact() {

	allOnes := uint64(0xffffffffffffffff)

	for len(tb.Words) > 0 && tb.Words[0] == allOnes {
		tb.Offset += 64
		tb.Words = tb.Words[1:]
	}

	if tb.Offset-tb.Reclamed >= reclaimThreshold {
		l := len(tb.Words)
		newWords := make([]uint64, l, l*2)

		copy(newWords, tb.Words)
		tb.Reclamed = tb.Offset
	}
}

// Set the bit at `idx` to `1`.
//
// Since 0.1.22
func (tb *TailBitmap) Set(idx int64) {
	if idx < tb.Offset {
		return
	}

	idx = idx - tb.Offset
	wordIdx := idx >> 6

	for int(wordIdx) >= len(tb.Words) {
		tb.Words = append(tb.Words, 0)
	}

	tb.Words[wordIdx] |= bitmap.Bit[idx&63]

	if wordIdx == 0 {
		tb.Compact()
	}
}

// Get retrieves a bit at its 64-based offset.
//
// Since 0.1.22
func (tb *TailBitmap) Get(idx int64) uint64 {
	if idx < tb.Offset {
		return bitmap.Bit[idx&63]
	}

	idx = idx - tb.Offset
	return tb.Words[idx>>6] & bitmap.Bit[idx&63]
}

// Get1 retrieves a bit and returns a 1-bit word, i.e., putting the bit in the
// lowest bit.
//
// Since 0.1.22
func (tb *TailBitmap) Get1(idx int64) uint64 {
	if idx < tb.Offset {
		return 1
	}
	idx = idx - tb.Offset
	return (tb.Words[idx>>6] >> uint(idx&63)) & 1
}

func (tb *TailBitmap) Union(tc *TailBitmap) {

	lb := tb.Offset + int64(len(tb.Words)*64)
	lc := tc.Offset + int64(len(tc.Words)*64)

	if tb.Offset >= lc {
		return
	}

	if lb <= tc.Offset {
		tb.Offset = tc.Offset
		tb.Words = make([]uint64, len(tc.Words))
		copy(tb.Words, tc.Words)

		// building a new Words reclames unused spaces in it.
		tb.Reclamed = tb.Offset
		return
	}

	var ws []uint64
	if tb.Offset >= tc.Offset {
		delta := tb.Offset - tc.Offset
		ws = tc.Words[delta>>6:]

	} else {
		// tb.Offset < tc.Offset

		delta := tc.Offset - tb.Offset
		tb.Words = tb.Words[delta>>6:]
		tb.Offset = tc.Offset
		ws = tc.Words
	}

	var i int
	for i = 0; i < len(tb.Words) && i < len(ws); i++ {
		tb.Words[i] |= ws[i]
	}

	for ; i < len(ws); i++ {
		tb.Words = append(tb.Words, ws[i])
	}

	tb.Compact()
}
