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

type logStat interface {
	GetCommitter() *LeaderId
	GetAccepted() *TailBitmap
}

type leaderStat interface {
	GetVotedFor() *LeaderId
	GetVoteExpireAt() int64
}

func CmpLogStatus(a, b logStat) int {
	r := a.GetCommitter().Cmp(b.GetCommitter())
	if r != 0 {
		return r
	}

	return cmpI64(a.GetAccepted().Len(), b.GetAccepted().Len())
}

func ExportLogStatus(ls logStat) *LogStatus {
	return &LogStatus{
		Committer: ls.GetCommitter().Clone(),
		Accepted:  ls.GetAccepted().Clone(),
	}
}

func ExportLeaderStatus(ls leaderStat) *LeaderStatus {
	return &LeaderStatus{
		VotedFor:     ls.GetVotedFor().Clone(),
		VoteExpireAt: ls.GetVoteExpireAt(),
	}
}

// CmpAccepted compares log related fields with another ballot.
// I.e. Committer and MaxLogSeq.
func (a *ReplicaStatus) CmpAccepted(b *ReplicaStatus) int {
	return CmpLogStatus(a, b)
}
