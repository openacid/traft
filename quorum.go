package traft

import "math/bits"

func buildMajorityQuorums(mask uint64) []uint64 {
	rst := make([]uint64, 0)
	major := bits.OnesCount64(mask)/2 + 1
	for i := uint64(0); i <= mask; i++ {
		if i&mask == i && bits.OnesCount64(i) == major {
			rst = append(rst, i)
		}
	}
	return rst
}
