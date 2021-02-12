package traft

// newStatusAcc creates a ReplicaStatus with only accepted fields inited.
// Mostly for test purpose only.
func newStatusAcc(aterm, aid, lsn int64) *ReplicaStatus {
	acc := NewTailBitmap((lsn + 1) &^ 63)
	if (lsn+1)&63 != 0 {
		acc.Words = append(acc.Words, 1<<uint(lsn&63))
	}
	return &ReplicaStatus{
		Committer: NewLeaderId(aterm, aid),
		Accepted:  acc,
	}
}

// CmpAccepted compares log related fields with another ballot.
// I.e. Committer and MaxLogSeq.
func (a *ReplicaStatus) CmpAccepted(b *ReplicaStatus) int {
	r := a.Committer.Cmp(b.Committer)
	if r != 0 {
		return r
	}

	return cmpI64(a.Accepted.Len(), b.Accepted.Len())
}
