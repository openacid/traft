package traft

import (
	context "context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/kr/pretty"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func cmpI64(a, b int64) int {
	if a > b {
		return 1
	}
	if a < b {
		return -1
	}
	return 0
}

var basePort = int64(5500)

// serveCluster starts a grpc server for every replica.
func serveCluster(ids []int64) ([]*grpc.Server, []*TRaft) {

	cluster := make(map[int64]string)

	servers := []*grpc.Server{}
	trafts := make([]*TRaft, 0)

	for _, id := range ids {
		addr := fmt.Sprintf(":%d", basePort+int64(id))
		cluster[id] = addr
	}

	for _, id := range ids {

		addr := cluster[id]

		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("listen: %s %v", addr, err)
		}

		node := NewNode(id, cluster)
		srv := &TRaft{Node: *node}
		trafts = append(trafts, srv)

		s := grpc.NewServer()
		RegisterTRaftServer(s, srv)
		reflection.Register(s)

		pretty.Logf("Acceptor-%d serving on %s ...", id, addr)
		servers = append(servers, s)
		go s.Serve(lis)
	}

	return servers, trafts
}

// send rpc to addr.
func rpcTo(addr string,
	action func(TRaftClient, context.Context)) {

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		// TODO check error
		panic("wooooooh")
	}
	defer conn.Close()

	cli := NewTRaftClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	action(cli, ctx)
}
