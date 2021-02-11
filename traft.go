package traft

func NewNode(ids []int64, addrs []string) *Node {
	members := make([]*ReplicaInfo, 0)
	for i, id := range ids {
		members = append(members, &ReplicaInfo{
			Id:   id,
			Addr: addrs[i],
		})
	}

	conf := &ClusterConfig{
		Members: members,
		Quorums: buildMajorityQuorums(1 << uint(len(members)-1)),
	}

	progs := make([]*ReplicaProgress, 0)
	for _, m := range members {
		_ = m
		progs = append(progs, &ReplicaProgress{
			// Has: &TailBitmap{},
		})
	}

	node := &Node{
		Config: conf,
		Log:    make([]*Record, 0),
		Term:   0,
		// AcceptedTerm: 0,
		// Progresses: progs,
	}

	return node
}
