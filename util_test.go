package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_serveCluster(t *testing.T) {

	ta := require.New(t)

	ids := []int64{1, 2, 3}

	servers, trafts := serveCluster(ids)

	defer func() {
		for _, s := range servers {
			s.Stop()
		}
	}()

	for i, tr := range trafts {
		ta.Equal(&ClusterConfig{
			Members: []*ReplicaInfo{
				&ReplicaInfo{Id: 1, Addr: ":5501"},
				&ReplicaInfo{Id: 2, Addr: ":5502"},
				&ReplicaInfo{Id: 3, Addr: ":5503"},
			},
			Quorums: []uint64{
				3, 5, 6,
			},
		}, tr.Config)

		ta.Equal(int64(0), tr.LogOffset)
		ta.Equal([]*Record{}, tr.Log)
		ta.Equal(ids[i], tr.Id)
		for _, id := range ids {
			st := tr.Status[id]
			ta.Equal(&ReplicaStatus{
				// voted for ones self at first.
				VotedFor:     &LeaderId{Term: 0, Id: id},
				AcceptedFrom: nil,
				Accepted:     NewTailBitmap(0),
				Committed:    NewTailBitmap(0),
				Applied:      NewTailBitmap(0),
			}, st)
		}
	}
}
