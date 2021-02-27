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


func emptyReplicaStatus(id int64) *ReplicaStatus {
	return &ReplicaStatus{
		// initially it votes for itself with term 0
		VotedFor:  NewLeaderId(0, id),
		Committer: nil,
		Accepted:  NewTailBitmap(0),
		Committed: NewTailBitmap(0),
		Applied:   NewTailBitmap(0),
	}
}


type logStater interface {
	GetCommitter() *LeaderId
	GetAccepted() *TailBitmap
}

func CmpLogStatus(a, b logStater) int {
	r := a.GetCommitter().Cmp(b.GetCommitter())
	if r != 0 {
		return r
	}

	return cmpI64(a.GetAccepted().Len(), b.GetAccepted().Len())
}

func ExportLogStatus(ls logStater) *LogStatus {
	return &LogStatus{
		Committer: ls.GetCommitter().Clone(),
		Accepted:  ls.GetAccepted().Clone(),
	}
}

// CmpAccepted compares log related fields with another ballot.
// I.e. Committer and MaxLogSeq.
func (a *ReplicaStatus) CmpAccepted(b *ReplicaStatus) int {
	return CmpLogStatus(a, b)
}

// If I have a log from a leader and the leader has committed it.
// I commit it too.
func (a *ReplicaStatus) UpdatedCommitted(committer *LeaderId, committed *TailBitmap) {
	if committer.Cmp(a.Committer) != 0 {
		panic("can not accept committed from non-leader committer")
	}

	update := a.Accepted.Clone()
	update.Intersection(committed)
	a.Committed.Union(update)
}
