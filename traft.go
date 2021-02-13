// Package traft is a raft variant with out-of-order commit/apply
// and a more generalized member change algo.
package traft

import "sort"

func NewNode(id int64, idAddrs map[int64]string) *Node {
	_, ok := idAddrs[id]
	if !ok {
		panic("my id is not in cluster")
	}

	members := make(map[int64]*ReplicaInfo, 0)

	ids := []int64{}
	for id, _ := range idAddrs {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	for p, id := range ids {
		members[id] = &ReplicaInfo{
			Id:       id,
			Addr:     idAddrs[id],
			Position: int64(p),
		}
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
