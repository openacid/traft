package traft

func NewBallot(cterm, cid, aterm, aid, lsn int64) *Ballot {
	return &Ballot{
		Current:      NewLeaderId(cterm, cid),
		MaxLogSeq:    lsn,
		AcceptedFrom: NewLeaderId(aterm, aid),
	}
}

// CmpLog compares log related fields with another ballot.
// I.e. AcceptedFrom and MaxLogSeq.
func (a *Ballot) CmpLog(b *Ballot) int {
	r := a.AcceptedFrom.Cmp(b.AcceptedFrom)
	if r != 0 {
		return r
	}

	return cmpI64(a.MaxLogSeq, b.MaxLogSeq)
}
