package traft

// Compare two leader id and returns 1, 0 or -1 for greater, equal and less
func (a *LeaderId) Cmp(b *LeaderId) int {
	r := cmpI64(a.Term, b.Term)
	if r != 0 {
		return r
	}

	return cmpI64(a.Id, b.Id)
}
