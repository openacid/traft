package traft

func NewBallot(cterm, cid, aterm, aid int64) *Ballot {
	return &Ballot{
		Current:  NewLeaderId(cterm, cid),
		Accepted: NewLeaderId(aterm, aid),
	}
}
