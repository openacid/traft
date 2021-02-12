package traft

// func NewBallot(cterm, cid, aterm, aid, lsn int64) *Ballot {
//     return &Ballot{
//         Current:      NewLeaderId(cterm, cid),
//         MaxLogSeq:    lsn,
//         Committer: NewLeaderId(aterm, aid),
//     }
// }

// // CmpLog compares log related fields with another ballot.
// // I.e. Committer and MaxLogSeq.
// func (a *Ballot) CmpLog(b *Ballot) int {
//     r := a.Committer.Cmp(b.Committer)
//     if r != 0 {
//         return r
//     }

//     return cmpI64(a.MaxLogSeq, b.MaxLogSeq)
// }
