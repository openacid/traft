package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClusterConfig_SortedReplicaInfos(t *testing.T) {

	ta := require.New(t)

	cc := &Cluster{
		Members: map[int64]*ReplicaInfo{
			1: {1, "111", 0},
			2: {2, "222", 2},
			3: {3, "333", 4},
		},
	}

	sorted := cc.SortedReplicaInfos()

	cases := []struct {
		input int64
		want  *ReplicaInfo
	}{
		{0, &ReplicaInfo{1, "111", 0}},
		{1, nil},
		{2, &ReplicaInfo{2, "222", 2}},
		{3, nil},
		{4, &ReplicaInfo{3, "333", 4}},
	}

	for i, c := range cases {
		got := sorted[c.input]
		ta.Equal(c.want, got, "%d-th: case: %+v", i+1, c)
	}
}

func TestClusterConfig_IsQuorum(t *testing.T) {

	ta := require.New(t)

	cc := &Cluster{
		Members: map[int64]*ReplicaInfo{
			1: {1, "111", 0},
			2: {2, "222", 2},
			3: {3, "333", 4},
		},
	}
	// quorums are:
	// 10100
	// 00101
	// 10001
	cc.Quorums = buildMajorityQuorums(1 | 4 | 16)

	cases := []struct {
		input uint64
		want  bool
	}{
		{1, false},
		{4, false},
		{16, false},
		{1 | 4, true},
		{4 | 16, true},
		{1 | 16, true},
		{1 | 4 | 16, true},
	}

	for i, c := range cases {
		got := cc.IsQuorum(c.input)
		ta.Equal(c.want, got, "%d-th: case: %+v", i+1, c)
	}
}
