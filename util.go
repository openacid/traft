package traft

func cmpI64(a, b int64) int {
	if a > b {
		return 1
	}
	if a < b {
		return -1
	}
	return 0
}
