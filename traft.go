// Package traft is a raft variant with out-of-order commit/apply
// and a more generalized member change algo.
package traft

import (
	fmt "fmt"
	"net"
	"sort"
	"strings"
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
		MsgCh:      make(chan string, 1024),
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
	tr.StartMainLoop()
	tr.StartVoteLoop()
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

func (tr *TRaft) goit(f func()) {
	tr.wg.Add(1)
	go func() {
		defer tr.wg.Done()
		f()
	}()
}

func (tr *TRaft) StartMainLoop() {
	tr.goit(tr.Loop)
	lg.Infow("Started Loop")
}

func (tr *TRaft) StartVoteLoop() {
	tr.goit(tr.VoteLoop)
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

func (tr *TRaft) sendMsg(msg ...interface{}) {
	type sstr interface {
		ShortStr() string
	}

	type str interface {
		String() string
	}

	mm := []string{fmt.Sprintf("Id=%d", tr.Id)}
	for _, m := range msg {
		{
			ss, ok := m.(sstr)
			if ok {
				mm = append(mm, ss.ShortStr())
				continue
			}
		}
		{
			ss, ok := m.(str)
			if ok {
				mm = append(mm, ss.String())
				continue
			}
		}
		mm = append(mm, fmt.Sprintf("%v", m))
	}

	vv := strings.Join(mm, " ")

	select {
	case tr.MsgCh <- vv:
		lg.Infow("sent msg", "msg", vv)
	default:
		lg.Infow("failure sending msg", "msg", vv)
	}
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
