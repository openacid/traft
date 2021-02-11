package traft

func (cc *ClusterConfig) GetReplicaInfo(id int64) *ReplicaInfo {
	for _, m := range cc.Members {
		if m.Id == id {
			return m
		}
	}
	return nil
}
