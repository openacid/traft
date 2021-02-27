package traft

import proto "github.com/gogo/protobuf/proto"

func (cc *Cluster) MaxPosition() int64 {
	maxPos := int64(0)
	for _, m := range cc.Members {
		if maxPos < m.Position {
			maxPos = m.Position
		}
	}

	return maxPos
}

func (cc *Cluster) Clone() *Cluster {
	return proto.Clone(cc).(*Cluster)
}

func (cc *Cluster) SortedReplicaInfos() []*ReplicaInfo {
	maxPos := cc.MaxPosition()

	members := make([]*ReplicaInfo, maxPos+1)

	for _, m := range cc.Members {
		members[m.Position] = m
	}

	return members
}

// check if a set of member is a quorum.
// The set of member is a bitmap in which a `1` indicates a present member.
// In this system, the position of `1` is ReplicaInfo.Position.
func (cc *Cluster) IsQuorum(v uint64) bool {

	for _, q := range cc.Quorums {
		if v&q == q {
			return true
		}
	}

	return false
}
