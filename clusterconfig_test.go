package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClusterConfig_GetReplicaInfo(t *testing.T) {

	ta := require.New(t)

	cc := &ClusterConfig{
		Members: []*ReplicaInfo{
			{1, "111"},
			{3, "333"},
			{2, "222"},
		},
	}

	cases := []struct {
		input int64
		want  *ReplicaInfo
	}{
		{0, nil},
		{1, &ReplicaInfo{1, "111"}},
		{2, &ReplicaInfo{2, "222"}},
		{3, &ReplicaInfo{3, "333"}},
		{4, nil},
	}

	for i, c := range cases {
		got := cc.GetReplicaInfo(c.input)
		ta.Equal(c.want, got, "%d-th: case: %+v", i+1, c)
	}
}
