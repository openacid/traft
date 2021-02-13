package traft

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_serveCluster(t *testing.T) {

	ta := require.New(t)

	ids := []int64{1, 2, 3}

	trafts := serveCluster(ids)

	defer func() {
		for _, s := range trafts {
			s.Stop()
		}
	}()

	for i, tr := range trafts {
		// 110 101 011
		ta.Equal([]uint64{3, 5, 6}, tr.Config.Quorums)

		fmt.Println("===", tr.Config.Members)
		fmt.Println("---", tr.Config.SortedReplicaInfos())

		ta.Equal([]*ReplicaInfo{
			&ReplicaInfo{Id: 1, Addr: ":5501", Position: 0},
			&ReplicaInfo{Id: 2, Addr: ":5502", Position: 1},
			&ReplicaInfo{Id: 3, Addr: ":5503", Position: 2},
		}, tr.Config.SortedReplicaInfos())

		ta.Equal(int64(0), tr.LogOffset)
		ta.Equal([]*Record{}, tr.Logs)
		ta.Equal(ids[i], tr.Id)
		for _, id := range ids {
			st := tr.Status[id]
			ta.Equal(&ReplicaStatus{
				// voted for ones self at first.
				VotedFor:  &LeaderId{Term: 0, Id: id},
				Committer: nil,
				Accepted:  NewTailBitmap(0),
				Committed: NewTailBitmap(0),
				Applied:   NewTailBitmap(0),
			}, st)
		}
	}
}
