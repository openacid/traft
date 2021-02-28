package traft

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var (
	meth = struct{
		LogForward, Vote, Propose string
	}{
		LogForward: "LogForward",
		Vote: "Vote",
		Propose: "Propose",
	}
)

// rpcResult is a container of rpc reply and other supporting info.
type rpcResult struct {
	addr   string
	method string

	reply interface{}
	err   error
}

// rpcSession is a session of RPCs to all members in a cluster except the sender.
type rpcSession struct {
	// context for all rpc
	ctx context.Context

	// call it if we have collected enough response and do not need to wait for
	// other rpc replies
	cancel context.CancelFunc

	// the cluster to send rpc to.
	// cluster must not be modified by other goroutine.
	cluster *Cluster

	// the method name, one of "Vote", "LogForward" and "Propose"
	method string

	// the request body
	req proto.Message

	// a channel to receive responded replies.
	resCh chan *rpcResult

	// bitmap of peers that responded positive reply, i.e., reply responded, and
	// the field "OK" is true.
	// The bit position for a peer is ReplicaInfo.Position
	okBitmap uint64

	// count of unresponded peers
	pending int64

	// whether a quorum is constitued, e.g., 3/5 positive replies received.
	quorum int32
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

func rpcToAll(
	id int64,
	cluster *Cluster,
	method string,
	req proto.Message,
	timeout time.Duration,
) *rpcSession {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	ms := cluster.Members

	sess := &rpcSession{
		ctx:    ctx,
		cancel: cancel,

		cluster: cluster,
		method:  method,
		req:     req,

		resCh: make(chan *rpcResult, len(ms)),

		okBitmap: 1 << uint(ms[id].Position),
		pending:  int64(len(ms) - 1),
		quorum:   0,
	}

	for _, m := range cluster.Members {
		if m.Id == id {
			continue
		}
		go func(ri ReplicaInfo) {
			res := rpcToPeer(ri, sess)
			sess.resCh <- res
		}(*m)
	}

	return sess
}

// rpcToPeer sends request and wait for the reply.
// It also update essential info such as:
// - pending: the N.O. unfinished rpcs.
// - okBitmap: a bitmap indicates which peer responded a reply with OK=true.
// - quorum: whether OK replies constitue a quorum.
func rpcToPeer(ri ReplicaInfo, sess *rpcSession) *rpcResult {

	res := &rpcResult{
		addr:   ri.Addr,
		method: sess.method,
		reply:  nil,
		err:    nil,
	}

	conn, err := grpc.Dial(ri.Addr, grpc.WithInsecure())
	if err != nil {
		lg.Infow("rpc-to", "addr", ri.Addr, "err", err)
		res.err = errors.Wrapf(err, "to %s", ri.Addr)
		return res
	}
	defer conn.Close()

	res.reply = newReply(sess.method)
	res.err = conn.Invoke(sess.ctx, "/TRaft/"+sess.method, sess.req, res.reply)

	// pending will be read by other goroutine thus must be read/written
	// atomically.
	atomic.AddInt64(&sess.pending, -1)

	if res.err != nil {
		lg.Infow("rpc-reply", "err", err)
		return res
	}

	type getOKer interface {
		GetOK() bool
	}
	if !res.reply.(getOKer).GetOK() {
		return res
	}

	// okBitmap will be read by other goroutine thus must be read/written
	// atomically.
	bm := casOrU64(&sess.okBitmap, uint64(1)<<uint(ri.Position))
	if sess.cluster.IsQuorum(bm) {
		atomic.StoreInt32(&sess.quorum, 1)
	}

	return res
}

// newReply creates an empty reply structure by method name.
// method name is one of the RPC func defined in traft.proto.
func newReply(method string) proto.Message {
	switch method {
	case "Vote":
		return &VoteReply{}
	case "LogForward":
		return &LogForwardReply{}
	case "Propose":
		return &ProposeReply{}
	default:
		panic("unknown method:" + method)
	}
}

// use check-and-swap loop to atomically set a bit in an uint64
func casOrU64(addr *uint64, mask uint64) uint64 {
	for {
		oldV := atomic.LoadUint64(addr)
		newV := oldV | mask
		if atomic.CompareAndSwapUint64(addr, oldV, newV) {
			return newV
		}
	}
}
