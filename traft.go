// Package traft is a raft variant with out-of-order commit/apply
// and a more generalized member change algo.
package traft

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type TRaft struct {
	running int64

	// close it to notify all goroutines to shutdown.
	shutdown chan struct{}

	// Communication channel with Loop().
	// Only Loop() modifies state of TRaft.
	// Other goroutines send an queryBody through this channel and wait for an
	// operation reply.
	actionCh chan *queryBody

	// for external component to receive traft state changes.
	MsgCh chan string

	grpcServer *grpc.Server

	// wait group of all worker goroutines
	workerWG sync.WaitGroup

	Node
}

func init() {
	initLogging()
}

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

	conf := &Cluster{
		Members: members,
	}
	maxPos := conf.MaxPosition()
	conf.Quorums = buildMajorityQuorums(1<<uint(maxPos+1) - 1)

	progs := make(map[int64]*ReplicaStatus, 0)
	for _, m := range members {
		progs[m.Id] = emptyReplicaStatus(m.Id)
	}

	node := &Node{
		Config: conf,
		Logs:   make([]*LogRecord, 0),
		Id:     id,
		Status: progs,
	}

	// TODO buffer size
	shutdown := make(chan struct{})
	actionCh := make(chan *queryBody)

	tr := &TRaft{
		running:    1,
		shutdown:   shutdown,
		actionCh:   actionCh,
		MsgCh:      make(chan string, 1024),
		grpcServer: nil,
		workerWG:   sync.WaitGroup{},
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
	tr.workerWG.Add(1)
	go func() {
		defer tr.workerWG.Done()
		f()
	}()
}

func (tr *TRaft) StartMainLoop() {
	tr.goit(tr.Loop)
	lg.Infow("Started Loop")
}

func (tr *TRaft) StartVoteLoop() {
	tr.goit(tr.ElectLoop)
	lg.Infow("Start ElectLoop")
}

func (tr *TRaft) Stop() {
	id := tr.Id
	addr := tr.Config.Members[id].Addr
	lg.Infow("Stopping grpc: ", "addr:", addr)
	// tr.grpcServer.Stop() does not wait.
	tr.grpcServer.GracefulStop()

	if atomic.LoadInt64(&tr.running) == 0 {
		lg.Infow("TRaft already stopped")
		return
	}

	lg.Infow("close shutdown")
	close(tr.shutdown)
	atomic.StoreInt64(&tr.running, 0)

	tr.workerWG.Wait()

	lg.Infow("TRaft stopped")
}

// stoppable sleep, if tr.Stop() has been called, it returns at once
func (tr *TRaft) sleep(t time.Duration) {
	select {
	case <-time.After(t):
	case <-tr.shutdown:
	}
}

func (tr *TRaft) sendMsg(msg ...interface{}) {

	mm := []string{fmt.Sprintf("Id=%d", tr.Id)}
	for _, m := range msg {
		mm = append(mm, toStr(m))
	}

	vv := strings.Join(mm, " ")
	fmt.Println("===", vv)

	select {
	case tr.MsgCh <- vv:
		// lg.Infow("succ-send-msg", "msg", vv)
	default:
		lg.Infow("fail-send-msg", "msg", vv)
	}
}
