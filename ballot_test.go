package traft

// func TestNewBallot(t *testing.T) {

//     ta := require.New(t)

//     got := NewBallot(1, 2, 3, 4, 5)
//     ta.Equal(int64(1), got.Current.Term)
//     ta.Equal(int64(2), got.Current.Id)

//     ta.Equal(int64(5), got.MaxLogSeq)

//     ta.Equal(int64(3), got.AcceptedFrom.Term)
//     ta.Equal(int64(4), got.AcceptedFrom.Id)
// }

// func TestBallog_CmpLog(t *testing.T) {

//     ta := require.New(t)

//     cases := []struct {
//         a, b *Ballot
//         want int
//     }{
//         {a: NewBallot(0, 0, 1, 1, 1), b: NewBallot(0, 0, 1, 1, 1), want: 0},
//         {a: NewBallot(0, 0, 1, 1, 2), b: NewBallot(0, 0, 1, 1, 1), want: 1},
//         {a: NewBallot(0, 0, 1, 2, 0), b: NewBallot(0, 0, 1, 1, 1), want: 1},
//         {a: NewBallot(0, 0, 2, 0, 0), b: NewBallot(0, 0, 1, 1, 1), want: 1},
//     }

//     for i, c := range cases {
//         ta.Equal(c.want, c.a.CmpLog(c.b), "%d-th: case: %+v", i+1, c)
//         ta.Equal(-c.want, c.b.CmpLog(c.a), "%d-th: case: %+v", i+1, c)
//     }
// }
