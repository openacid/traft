package traft

func (cc *ClusterConfig) MaxPosition() int64 {
	maxPos := int64(0)
	for _, m := range cc.Members {
		if maxPos < m.Position {
			maxPos = m.Position
		}
	}

	return maxPos
}

func (cc *ClusterConfig) SortedReplicaInfos() []*ReplicaInfo {
	maxPos := cc.MaxPosition()

	members := make([]*ReplicaInfo, maxPos+1)

	for _, m := range cc.Members {
		members[m.Position] = m
	}

	return members
}
