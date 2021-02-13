// Package traft is a raft variant with out-of-order commit/apply
// and a more generalized member change algo.
package traft

import (
	"net"
	"sort"
	sync "sync"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewTRaft(id int64, idAddrs map[int64]string) *TRaft {
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

	// TODO buffer size
	shutdown := make(chan struct{})
	actionCh := make(chan *action)

	tr := &TRaft{
		shutdown:   shutdown,
		actionCh:   actionCh,
		grpcServer: nil,
		wg:         sync.WaitGroup{},
		Node:       *node,
	}

	{
		s := grpc.NewServer()
		RegisterTRaftServer(s, tr)
		reflection.Register(s)

		tr.grpcServer = s
	}

	return tr
}

func (tr *TRaft) Start() {
	tr.StartServer()
	tr.StartLoops()
}

func (tr *TRaft) StartServer() {

	id := tr.Id
	addr := tr.Config.Members[id].Addr

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		lg.Fatalw("Fail to listen:", "addr", addr, "err", err)
	}

	go tr.grpcServer.Serve(lis)
	lg.Infow("grpc started", "addr", addr)
}

func (tr *TRaft) StartLoops() {

	tr.wg.Add(1)
	go tr.Loop(tr.shutdown, tr.actionCh)
	lg.Infow("Started Loop")

	tr.wg.Add(1)
	go tr.VoteLoop(tr.shutdown, tr.actionCh)
	lg.Infow("Start VoteLoop")
}

func (tr *TRaft) Stop() {
	lg.Infow("Stopping grpc")
	tr.grpcServer.Stop()

	lg.Infow("close shutdown")
	close(tr.shutdown)

	tr.wg.Wait()
	lg.Infow("TRaft stopped")
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
