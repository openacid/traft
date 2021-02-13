// Package traft is a raft variant with out-of-order commit/apply
// and a more generalized member change algo.
package traft

func NewNode(id int64, idAddrs map[int64]string) *Node {
	_, ok := idAddrs[id]
	if !ok {
		panic("my id is not in cluster")
	}

	members := make(map[int64]*ReplicaInfo, 0)
	p := int64(0)
	for id, addr := range idAddrs {
		members[id] = &ReplicaInfo{
			Id:       id,
			Addr:     addr,
			Position: p,
		}
		p++
	}

	conf := &ClusterConfig{
		Members: members,
	}
	maxPos := conf.MaxPosition()
	conf.Quorums = buildMajorityQuorums(1<<uint(maxPos+1) - 1)

	progs := make(map[int64]*ReplicaStatus, 0)
	for _, m := range members {
		progs[m.Id] = emptyProgress(m.Id)
	}

	node := &Node{
		Config: conf,
		Logs:   make([]*Record, 0),
		Id:     id,
		Status: progs,
	}

	return node
}

func emptyProgress(id int64) *ReplicaStatus {
	return &ReplicaStatus{
		// initially it votes for itself with term 0
		VotedFor:  NewLeaderId(0, id),
		Committer: nil,
		Accepted:  NewTailBitmap(0),
		Committed: NewTailBitmap(0),
		Applied:   NewTailBitmap(0),
	}
}
