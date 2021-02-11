package traft

func NewNode(id int64, idAddrs map[int64]string) *Node {
	_, ok := idAddrs[id]
	if !ok {
		panic("my id is not in cluster")
	}

	members := make([]*ReplicaInfo, 0)
	for id, addr := range idAddrs {
		members = append(members, &ReplicaInfo{
			Id:   id,
			Addr: addr,
		})
	}

	conf := &ClusterConfig{
		Members: members,
		Quorums: buildMajorityQuorums(1 << uint(len(members)-1)),
	}

	progs := make(map[int64]*ReplicaProgress, 0)
	for _, m := range members {
		progs[m.Id] = emptyProgress()
	}

	node := &Node{
		Config:     conf,
		Log:        make([]*Record, 0),
		Term:       0,
		Id:         id,
		VotedFor:   nil,
		Progresses: progs,
	}

	return node
}

func emptyProgress() *ReplicaProgress {
	return &ReplicaProgress{
		AcceptedFrom: nil,
		Accepted:     NewTailBitmap(0),
		Committed:    NewTailBitmap(0),
		Applied:      NewTailBitmap(0),
	}

}
