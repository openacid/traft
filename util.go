package traft

import (
	context "context"
	"fmt"
	"strings"
	"time"

	grpc "google.golang.org/grpc"
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

func uSecondI64() int64 {
	return int64(uSecond())
}

func uSecond() time.Duration {
	now := time.Now()
	return time.Duration(now.Unix())*1000*1000*1000 + time.Duration(now.Nanosecond())
}

func join(ss ...string) string {
	return strings.Join(ss, "")
}

var basePort = int64(5500)

// serveCluster starts a grpc server for every replica.
func serveCluster(ids []int64) []*TRaft {

	cluster := make(map[int64]string)

	trafts := make([]*TRaft, 0)

	for _, id := range ids {
		addr := fmt.Sprintf(":%d", basePort+int64(id))
		cluster[id] = addr
	}

	for _, id := range ids {
		srv := NewTRaft(id, cluster)
		trafts = append(trafts, srv)

		// in a test env, only start server
		// manually start loops
		srv.StartServer()
		srv.StartMainLoop()
	}

	return trafts
}

// send rpc to addr.
// TODO use a single loop to send to one replica
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
