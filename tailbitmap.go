package traft

import (
	fmt "fmt"
	"math/bits"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	"github.com/openacid/low/bitmap"
	"github.com/openacid/low/mathext/util"
)

// reclaimThreshold is the size threshold in bit for reclamation of `Words`.
var reclaimThreshold = int64(1024) * 64

// NewTailBitmap creates an TailBitmap with a preset Offset and an empty
// tail bitmap.
//
// Optional arg `set` specifies what bit to set to 1.
// The bit positions in `set` is absolute, NOT based on offset.
//
// Since 0.1.22
func NewTailBitmap(offset int64, set ...int64) *TailBitmap {
	residual := offset & 63
	tb := &TailBitmap{
		Offset:   offset & ^63,
		Reclamed: offset & ^63,
		Words:    make([]uint64, 0, reclaimThreshold>>6),
	}
	if residual != 0 {
		for i := int64(0); i < residual; i++ {
			tb.Set(tb.Offset + i)
		}
	}
	for _, pos := range set {
		tb.Set(pos)
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
	if int(idx>>6) >= len(tb.Words) {
		return 0
	}
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

func (tb *TailBitmap) Clone() *TailBitmap {
	return proto.Clone(tb).(*TailBitmap)
}

func (tb *TailBitmap) Normalize() *TailBitmap {
	if tb.Words == nil {
		tb.Words = make([]uint64, 0)
	}
	return tb
}

func (tb *TailBitmap) Union(tc *TailBitmap) {

	if tc == nil {
		return
	}

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

func (ta *TailBitmap) Intersection(tb *TailBitmap) {

	if tb == nil {
		ta.Offset = 0
		ta.Words = make([]uint64, 0)
		ta.Reclamed = 0
		return
	}

	la := ta.Offset + int64(len(ta.Words)*64)
	lb := tb.Offset + int64(len(tb.Words)*64)

	// 1111 1111 xxxx
	// 1111 yyyy
	if ta.Offset >= lb {
		ta.Offset = tb.Offset
		ta.Words = make([]uint64, len(tb.Words))
		copy(ta.Words, tb.Words)

		// building a new Words reclames unused spaces in it.
		ta.Reclamed = ta.Offset
		return
	}

	// 1111 xxxx
	// 1111 1111 yyyy
	if la <= tb.Offset {
		return
	}

	s := util.MinI64(ta.Offset, tb.Offset)
	e := util.MinI64(la, lb)
	ws := make([]uint64, (e-s)>>6)
	cur := int64(0)
	i := int64(0)
	j := int64(0)
	if ta.Offset >= tb.Offset {
		n := (ta.Offset - s) >> 6
		n = util.MinI64(n, (e-s)>>6)
		copy(ws, tb.Words[:n])
		cur += n
		j = n
	} else {
		n := (tb.Offset - s) >> 6
		n = util.MinI64(n, (e-s)>>6)
		copy(ws, ta.Words[:n])
		cur += n
		i = n
	}

	for cur < int64(len(ws)) {
		ws[cur] = ta.Words[i] & tb.Words[j]
		cur++
		i++
		j++
	}

	for len(ws) > 0 && ws[len(ws)-1] == 0 {
		ws = ws[:len(ws)-1]
	}
	ta.Offset = s
	ta.Words = ws
	ta.Reclamed = s
}

// Diff AKA substraction A - B or  A \ B
// TODO: This impl is wrong!!!
func (tb *TailBitmap) Diff(tc *TailBitmap) {

	if tc == nil {
		return
	}

	lb := tb.Offset + int64(len(tb.Words)*64)
	lc := tc.Offset + int64(len(tc.Words)*64)

	if lb <= tc.Offset {
		for i := 0; i < len(tb.Words); i++ {
			tb.Words[i] = ^tb.Words[i]
		}
		return
	}

	if tb.Offset > lc {
		// 1111 1111 1111 xxxx xxxx
		// 1111 yyyy
		l := int((tb.Offset - tc.Offset) >> 6)
		words := make([]uint64, l+len(tb.Words))
		var i int
		for i = 0; i < l && i < len(tc.Words); i++ {
			words[i] = ^tc.Words[i]
		}
		for ; i < l; i++ {
			words[i] = 0xffffffffffffffff
		}

		copy(words[i:], tb.Words)
		tb.Words = words
		tb.Offset = tc.Offset
		tb.Reclamed = tb.Offset
		return
	}

	if tb.Offset <= tc.Offset {
		// 1111 1111 xxxx xxxx
		// 1111 1111 1111 yyyy yyyy
		delta := (tc.Offset - tb.Offset) >> 6
		var i int64
		for i = 0; i < delta; i++ {
			tb.Words[i] = 0
		}
		for ; i < int64(len(tb.Words)) && i < (lc-tb.Offset)>>6; i++ {
			tb.Words[i] &= ^tc.Words[i-delta]
		}

	} else {
		// tb.Offset > tc.Offset
		// 1111 1111 xxxx xxxx
		// 1111 yyyy yyyy

		delta := int((tb.Offset - tc.Offset) >> 6)
		words := make([]uint64, delta+len(tb.Words))

		var i int
		for i = 0; i < delta; i++ {
			words[i] = ^tc.Words[i]
		}
		for ; i < len(words) && i < len(tc.Words); i++ {
			words[i] = tb.Words[i-delta] &^ tc.Words[i]
		}
		copy(words[i:], tb.Words[i-delta:])

		tb.Words = words
		tb.Offset = tc.Offset
		tb.Reclamed = tb.Offset
	}
}

// Last returns last set bit index + 1.
func (tb *TailBitmap) Len() int64 {

	r := len(tb.Words) - 1
	for ; r >= 0 && tb.Words[r] == 0; r-- {
	}

	if r < 0 {
		// all Words are 0
		return tb.Offset
	}

	return tb.Offset + int64(r+1)<<6 - int64(bits.LeadingZeros64(tb.Words[r]))
}

func (tb *TailBitmap) ShortStr() string {
	if tb == nil {
		return "0"
	}
	s := []string{fmt.Sprintf("%d", tb.Offset)}
	for _, w := range tb.Words {
		s = append(s, fmt.Sprintf(":%x", w))
	}

	return strings.Join(s, "")
}

func (tb *TailBitmap) DebugStr() string {
	if tb == nil {
		return "0"
	}
	s := []string{fmt.Sprintf("%d", tb.Offset)}
	for _, w := range tb.Words {
		v := bitmap.Fmt(w)
		s = append(s, v)
	}

	return strings.Join(s, ",")
}
