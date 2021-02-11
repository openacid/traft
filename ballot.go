package traft

func NewBallot(cterm, cid, aterm, aid, lsn int64) *Ballot {
	return &Ballot{
		Current:   NewLeaderId(cterm, cid),
		MaxLogSeq: lsn,
		Accepted:  NewLeaderId(aterm, aid),
	}
}

// CmpLog compares log related fields with another ballot.
// I.e. Accepted and MaxLogSeq.
func (a *Ballot) CmpLog(b *Ballot) int {
	r := a.Accepted.Cmp(b.Accepted)
	if r != 0 {
		return r
	}

	return cmpI64(a.MaxLogSeq, b.MaxLogSeq)
}
