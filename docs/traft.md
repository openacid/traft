# traft
--
    import "github.com/openacid/traft"

Package traft is a raft variant with out-of-order commit/apply and a more
generalized member change algo.

## Usage

```go
var (
	ErrInvalidLengthTraft        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTraft          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTraft = fmt.Errorf("proto: unexpected end of group")
)
```

#### func  CmpLogStatus

```go
func CmpLogStatus(a, b logStat) int
```

#### func  RecordsShortStr

```go
func RecordsShortStr(rs []*Record) string
```

#### func  RegisterTRaftServer

```go
func RegisterTRaftServer(s *grpc.Server, srv TRaftServer)
```

#### type ClusterConfig

```go
type ClusterConfig struct {
	Members []*ReplicaInfo `protobuf:"bytes,11,rep,name=Members,proto3" json:"Members,omitempty"`
	Quorums []uint64       `protobuf:"varint,21,rep,packed,name=Quorums,proto3" json:"Quorums,omitempty"`
}
```


#### func (*ClusterConfig) Descriptor

```go
func (*ClusterConfig) Descriptor() ([]byte, []int)
```

#### func (*ClusterConfig) Equal

```go
func (this *ClusterConfig) Equal(that interface{}) bool
```

#### func (*ClusterConfig) GetMembers

```go
func (m *ClusterConfig) GetMembers() []*ReplicaInfo
```

#### func (*ClusterConfig) GetQuorums

```go
func (m *ClusterConfig) GetQuorums() []uint64
```

#### func (*ClusterConfig) GetReplicaInfo

```go
func (cc *ClusterConfig) GetReplicaInfo(id int64) *ReplicaInfo
```

#### func (*ClusterConfig) Marshal

```go
func (m *ClusterConfig) Marshal() (dAtA []byte, err error)
```

#### func (*ClusterConfig) MarshalTo

```go
func (m *ClusterConfig) MarshalTo(dAtA []byte) (int, error)
```

#### func (*ClusterConfig) MarshalToSizedBuffer

```go
func (m *ClusterConfig) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*ClusterConfig) ProtoMessage

```go
func (*ClusterConfig) ProtoMessage()
```

#### func (*ClusterConfig) Reset

```go
func (m *ClusterConfig) Reset()
```

#### func (*ClusterConfig) Size

```go
func (m *ClusterConfig) Size() (n int)
```

#### func (*ClusterConfig) String

```go
func (m *ClusterConfig) String() string
```

#### func (*ClusterConfig) Unmarshal

```go
func (m *ClusterConfig) Unmarshal(dAtA []byte) error
```

#### func (*ClusterConfig) XXX_DiscardUnknown

```go
func (m *ClusterConfig) XXX_DiscardUnknown()
```

#### func (*ClusterConfig) XXX_Marshal

```go
func (m *ClusterConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*ClusterConfig) XXX_Merge

```go
func (m *ClusterConfig) XXX_Merge(src proto.Message)
```

#### func (*ClusterConfig) XXX_Size

```go
func (m *ClusterConfig) XXX_Size() int
```

#### func (*ClusterConfig) XXX_Unmarshal

```go
func (m *ClusterConfig) XXX_Unmarshal(b []byte) error
```

#### type Cmd

```go
type Cmd struct {
	Op  string `protobuf:"bytes,10,opt,name=Op,proto3" json:"Op,omitempty"`
	Key string `protobuf:"bytes,20,opt,name=Key,proto3" json:"Key,omitempty"`
	// Types that are valid to be assigned to Value:
	//	*Cmd_VStr
	//	*Cmd_VI64
	//	*Cmd_VClusterConfig
	Value isCmd_Value `protobuf_oneof:"Value"`
}
```

Cmd defines the action a log record does

#### func  NewCmdI64

```go
func NewCmdI64(op, key string, v int64) *Cmd
```

#### func (*Cmd) Descriptor

```go
func (*Cmd) Descriptor() ([]byte, []int)
```

#### func (*Cmd) Equal

```go
func (this *Cmd) Equal(that interface{}) bool
```

#### func (*Cmd) GetKey

```go
func (m *Cmd) GetKey() string
```

#### func (*Cmd) GetOp

```go
func (m *Cmd) GetOp() string
```

#### func (*Cmd) GetVClusterConfig

```go
func (m *Cmd) GetVClusterConfig() *ClusterConfig
```

#### func (*Cmd) GetVI64

```go
func (m *Cmd) GetVI64() int64
```

#### func (*Cmd) GetVStr

```go
func (m *Cmd) GetVStr() string
```

#### func (*Cmd) GetValue

```go
func (m *Cmd) GetValue() isCmd_Value
```

#### func (*Cmd) Marshal

```go
func (m *Cmd) Marshal() (dAtA []byte, err error)
```

#### func (*Cmd) MarshalTo

```go
func (m *Cmd) MarshalTo(dAtA []byte) (int, error)
```

#### func (*Cmd) MarshalToSizedBuffer

```go
func (m *Cmd) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*Cmd) ProtoMessage

```go
func (*Cmd) ProtoMessage()
```

#### func (*Cmd) Reset

```go
func (m *Cmd) Reset()
```

#### func (*Cmd) ShortStr

```go
func (c *Cmd) ShortStr() string
```

#### func (*Cmd) Size

```go
func (m *Cmd) Size() (n int)
```

#### func (*Cmd) String

```go
func (m *Cmd) String() string
```

#### func (*Cmd) Unmarshal

```go
func (m *Cmd) Unmarshal(dAtA []byte) error
```

#### func (*Cmd) XXX_DiscardUnknown

```go
func (m *Cmd) XXX_DiscardUnknown()
```

#### func (*Cmd) XXX_Marshal

```go
func (m *Cmd) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*Cmd) XXX_Merge

```go
func (m *Cmd) XXX_Merge(src proto.Message)
```

#### func (*Cmd) XXX_OneofWrappers

```go
func (*Cmd) XXX_OneofWrappers() []interface{}
```
XXX_OneofWrappers is for the internal use of the proto package.

#### func (*Cmd) XXX_Size

```go
func (m *Cmd) XXX_Size() int
```

#### func (*Cmd) XXX_Unmarshal

```go
func (m *Cmd) XXX_Unmarshal(b []byte) error
```

#### type Cmd_VClusterConfig

```go
type Cmd_VClusterConfig struct {
	VClusterConfig *ClusterConfig `protobuf:"bytes,33,opt,name=VClusterConfig,proto3,oneof" json:"VClusterConfig,omitempty"`
}
```


#### func (*Cmd_VClusterConfig) Equal

```go
func (this *Cmd_VClusterConfig) Equal(that interface{}) bool
```

#### func (*Cmd_VClusterConfig) MarshalTo

```go
func (m *Cmd_VClusterConfig) MarshalTo(dAtA []byte) (int, error)
```

#### func (*Cmd_VClusterConfig) MarshalToSizedBuffer

```go
func (m *Cmd_VClusterConfig) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*Cmd_VClusterConfig) Size

```go
func (m *Cmd_VClusterConfig) Size() (n int)
```

#### type Cmd_VI64

```go
type Cmd_VI64 struct {
	VI64 int64 `protobuf:"varint,32,opt,name=VI64,proto3,oneof" json:"VI64,omitempty"`
}
```


#### func (*Cmd_VI64) Equal

```go
func (this *Cmd_VI64) Equal(that interface{}) bool
```

#### func (*Cmd_VI64) MarshalTo

```go
func (m *Cmd_VI64) MarshalTo(dAtA []byte) (int, error)
```

#### func (*Cmd_VI64) MarshalToSizedBuffer

```go
func (m *Cmd_VI64) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*Cmd_VI64) Size

```go
func (m *Cmd_VI64) Size() (n int)
```

#### type Cmd_VStr

```go
type Cmd_VStr struct {
	VStr string `protobuf:"bytes,31,opt,name=VStr,proto3,oneof" json:"VStr,omitempty"`
}
```


#### func (*Cmd_VStr) Equal

```go
func (this *Cmd_VStr) Equal(that interface{}) bool
```

#### func (*Cmd_VStr) MarshalTo

```go
func (m *Cmd_VStr) MarshalTo(dAtA []byte) (int, error)
```

#### func (*Cmd_VStr) MarshalToSizedBuffer

```go
func (m *Cmd_VStr) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*Cmd_VStr) Size

```go
func (m *Cmd_VStr) Size() (n int)
```

#### type LeaderId

```go
type LeaderId struct {
	Term int64 `protobuf:"varint,1,opt,name=Term,proto3" json:"Term,omitempty"`
	Id   int64 `protobuf:"varint,2,opt,name=Id,proto3" json:"Id,omitempty"`
}
```


#### func  NewLeaderId

```go
func NewLeaderId(term, id int64) *LeaderId
```

#### func (*LeaderId) Clone

```go
func (l *LeaderId) Clone() *LeaderId
```

#### func (*LeaderId) Cmp

```go
func (a *LeaderId) Cmp(b *LeaderId) int
```
Compare two leader id and returns 1, 0 or -1 for greater, equal and less

#### func (*LeaderId) Descriptor

```go
func (*LeaderId) Descriptor() ([]byte, []int)
```

#### func (*LeaderId) Equal

```go
func (this *LeaderId) Equal(that interface{}) bool
```

#### func (*LeaderId) GetId

```go
func (m *LeaderId) GetId() int64
```

#### func (*LeaderId) GetTerm

```go
func (m *LeaderId) GetTerm() int64
```

#### func (*LeaderId) Marshal

```go
func (m *LeaderId) Marshal() (dAtA []byte, err error)
```

#### func (*LeaderId) MarshalTo

```go
func (m *LeaderId) MarshalTo(dAtA []byte) (int, error)
```

#### func (*LeaderId) MarshalToSizedBuffer

```go
func (m *LeaderId) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*LeaderId) ProtoMessage

```go
func (*LeaderId) ProtoMessage()
```

#### func (*LeaderId) Reset

```go
func (m *LeaderId) Reset()
```

#### func (*LeaderId) ShortStr

```go
func (l *LeaderId) ShortStr() string
```

#### func (*LeaderId) Size

```go
func (m *LeaderId) Size() (n int)
```

#### func (*LeaderId) String

```go
func (m *LeaderId) String() string
```

#### func (*LeaderId) Unmarshal

```go
func (m *LeaderId) Unmarshal(dAtA []byte) error
```

#### func (*LeaderId) XXX_DiscardUnknown

```go
func (m *LeaderId) XXX_DiscardUnknown()
```

#### func (*LeaderId) XXX_Marshal

```go
func (m *LeaderId) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*LeaderId) XXX_Merge

```go
func (m *LeaderId) XXX_Merge(src proto.Message)
```

#### func (*LeaderId) XXX_Size

```go
func (m *LeaderId) XXX_Size() int
```

#### func (*LeaderId) XXX_Unmarshal

```go
func (m *LeaderId) XXX_Unmarshal(b []byte) error
```

#### type Node

```go
type Node struct {
	// replica id of this replica.
	Id     int64          `protobuf:"varint,3,opt,name=Id,proto3" json:"Id,omitempty"`
	Config *ClusterConfig `protobuf:"bytes,1,opt,name=Config,proto3" json:"Config,omitempty"`
	// From which log seq number we keeps here.
	LogOffset int64     `protobuf:"varint,4,opt,name=LogOffset,proto3" json:"LogOffset,omitempty"`
	Log       []*Record `protobuf:"bytes,2,rep,name=Log,proto3" json:"Log,omitempty"`
	// local view of every replica, including this node too.
	Status map[int64]*ReplicaStatus `protobuf:"bytes,6,rep,name=Status,proto3" json:"Status,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}
```


#### func  NewNode

```go
func NewNode(id int64, idAddrs map[int64]string) *Node
```

#### func (*Node) Descriptor

```go
func (*Node) Descriptor() ([]byte, []int)
```

#### func (*Node) Equal

```go
func (this *Node) Equal(that interface{}) bool
```

#### func (*Node) GetConfig

```go
func (m *Node) GetConfig() *ClusterConfig
```

#### func (*Node) GetId

```go
func (m *Node) GetId() int64
```

#### func (*Node) GetLog

```go
func (m *Node) GetLog() []*Record
```

#### func (*Node) GetLogOffset

```go
func (m *Node) GetLogOffset() int64
```

#### func (*Node) GetStatus

```go
func (m *Node) GetStatus() map[int64]*ReplicaStatus
```

#### func (*Node) Marshal

```go
func (m *Node) Marshal() (dAtA []byte, err error)
```

#### func (*Node) MarshalTo

```go
func (m *Node) MarshalTo(dAtA []byte) (int, error)
```

#### func (*Node) MarshalToSizedBuffer

```go
func (m *Node) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*Node) ProtoMessage

```go
func (*Node) ProtoMessage()
```

#### func (*Node) Reset

```go
func (m *Node) Reset()
```

#### func (*Node) Size

```go
func (m *Node) Size() (n int)
```

#### func (*Node) String

```go
func (m *Node) String() string
```

#### func (*Node) Unmarshal

```go
func (m *Node) Unmarshal(dAtA []byte) error
```

#### func (*Node) XXX_DiscardUnknown

```go
func (m *Node) XXX_DiscardUnknown()
```

#### func (*Node) XXX_Marshal

```go
func (m *Node) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*Node) XXX_Merge

```go
func (m *Node) XXX_Merge(src proto.Message)
```

#### func (*Node) XXX_Size

```go
func (m *Node) XXX_Size() int
```

#### func (*Node) XXX_Unmarshal

```go
func (m *Node) XXX_Unmarshal(b []byte) error
```

#### type Record

```go
type Record struct {
	// Which leader initially proposed this log.
	// Author may not be the same with Committer, if Author fails when trying to
	// commit a log record.
	//
	// TODO It seems this field is useless. Because we already have `Accepted`.
	// This is different from the original raft:
	// raft does not have a explicit concept `accepted`, which is enssential in
	// paxos.
	// Instead, The `commited` in raft is defined as: leader forwards its
	// own term log to a quorum.
	Author *LeaderId `protobuf:"bytes,1,opt,name=Author,proto3" json:"Author,omitempty"`
	// Log sequence number.
	Seq int64 `protobuf:"varint,10,opt,name=Seq,proto3" json:"Seq,omitempty"`
	// Cmd describes what this log does.
	Cmd *Cmd `protobuf:"bytes,30,opt,name=Cmd,proto3" json:"Cmd,omitempty"`
	// Overrides describes what previous logs this log record overrides.
	Overrides *TailBitmap `protobuf:"bytes,40,opt,name=Overrides,proto3" json:"Overrides,omitempty"`
}
```

Record is a log record

#### func  NewRecord

```go
func NewRecord(leader *LeaderId, seq int64, cmd *Cmd) *Record
```
NewRecord: without Overrides yet!!! TODO

#### func (*Record) Descriptor

```go
func (*Record) Descriptor() ([]byte, []int)
```

#### func (*Record) Equal

```go
func (this *Record) Equal(that interface{}) bool
```

#### func (*Record) GetAuthor

```go
func (m *Record) GetAuthor() *LeaderId
```

#### func (*Record) GetCmd

```go
func (m *Record) GetCmd() *Cmd
```

#### func (*Record) GetOverrides

```go
func (m *Record) GetOverrides() *TailBitmap
```

#### func (*Record) GetSeq

```go
func (m *Record) GetSeq() int64
```

#### func (*Record) Marshal

```go
func (m *Record) Marshal() (dAtA []byte, err error)
```

#### func (*Record) MarshalTo

```go
func (m *Record) MarshalTo(dAtA []byte) (int, error)
```

#### func (*Record) MarshalToSizedBuffer

```go
func (m *Record) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*Record) ProtoMessage

```go
func (*Record) ProtoMessage()
```

#### func (*Record) Reset

```go
func (m *Record) Reset()
```

#### func (*Record) ShortStr

```go
func (r *Record) ShortStr() string
```

#### func (*Record) Size

```go
func (m *Record) Size() (n int)
```

#### func (*Record) String

```go
func (m *Record) String() string
```

#### func (*Record) Unmarshal

```go
func (m *Record) Unmarshal(dAtA []byte) error
```

#### func (*Record) XXX_DiscardUnknown

```go
func (m *Record) XXX_DiscardUnknown()
```

#### func (*Record) XXX_Marshal

```go
func (m *Record) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*Record) XXX_Merge

```go
func (m *Record) XXX_Merge(src proto.Message)
```

#### func (*Record) XXX_Size

```go
func (m *Record) XXX_Size() int
```

#### func (*Record) XXX_Unmarshal

```go
func (m *Record) XXX_Unmarshal(b []byte) error
```

#### type ReplicaInfo

```go
type ReplicaInfo struct {
	Id   int64  `protobuf:"varint,1,opt,name=Id,proto3" json:"Id,omitempty"`
	Addr string `protobuf:"bytes,2,opt,name=Addr,proto3" json:"Addr,omitempty"`
}
```


#### func (*ReplicaInfo) Descriptor

```go
func (*ReplicaInfo) Descriptor() ([]byte, []int)
```

#### func (*ReplicaInfo) Equal

```go
func (this *ReplicaInfo) Equal(that interface{}) bool
```

#### func (*ReplicaInfo) GetAddr

```go
func (m *ReplicaInfo) GetAddr() string
```

#### func (*ReplicaInfo) GetId

```go
func (m *ReplicaInfo) GetId() int64
```

#### func (*ReplicaInfo) Marshal

```go
func (m *ReplicaInfo) Marshal() (dAtA []byte, err error)
```

#### func (*ReplicaInfo) MarshalTo

```go
func (m *ReplicaInfo) MarshalTo(dAtA []byte) (int, error)
```

#### func (*ReplicaInfo) MarshalToSizedBuffer

```go
func (m *ReplicaInfo) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*ReplicaInfo) ProtoMessage

```go
func (*ReplicaInfo) ProtoMessage()
```

#### func (*ReplicaInfo) Reset

```go
func (m *ReplicaInfo) Reset()
```

#### func (*ReplicaInfo) Size

```go
func (m *ReplicaInfo) Size() (n int)
```

#### func (*ReplicaInfo) String

```go
func (m *ReplicaInfo) String() string
```

#### func (*ReplicaInfo) Unmarshal

```go
func (m *ReplicaInfo) Unmarshal(dAtA []byte) error
```

#### func (*ReplicaInfo) XXX_DiscardUnknown

```go
func (m *ReplicaInfo) XXX_DiscardUnknown()
```

#### func (*ReplicaInfo) XXX_Marshal

```go
func (m *ReplicaInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*ReplicaInfo) XXX_Merge

```go
func (m *ReplicaInfo) XXX_Merge(src proto.Message)
```

#### func (*ReplicaInfo) XXX_Size

```go
func (m *ReplicaInfo) XXX_Size() int
```

#### func (*ReplicaInfo) XXX_Unmarshal

```go
func (m *ReplicaInfo) XXX_Unmarshal(b []byte) error
```

#### type ReplicaStatus

```go
type ReplicaStatus struct {
	// last seen term+id
	// int64 Term = 3;
	// int64 Id = 10;
	// the last leader it voted for. or it is local term + local id.
	// E.g., voted for itself.
	//
	// TODO cleanup comment:
	// which replica it has voted for as a leader.
	//
	// Accepted is the same as VotedFor after receiving one log-replication
	// message from the leader.
	//
	// Before receiving a message, VotedFor is the leader this replica knows of,
	// Accepted is nil.
	VotedFor *LeaderId `protobuf:"bytes,10,opt,name=VotedFor,proto3" json:"VotedFor,omitempty"`
	// The Leader tried to commit all of the local logs.
	// The Committer is the same as Author if a log entry is committed by its
	// Author.
	//
	// If an Author fails and the log is finally committed by some other leader,
	// Committer is a higher value than Author.
	//
	// It is similar to the vrnd/vballot concept in paxos.
	// the Ballot number a value is accepted at.
	Committer *LeaderId `protobuf:"bytes,4,opt,name=Committer,proto3" json:"Committer,omitempty"`
	// What logs has been accepted by this replica.
	Accepted  *TailBitmap `protobuf:"bytes,1,opt,name=Accepted,proto3" json:"Accepted,omitempty"`
	Committed *TailBitmap `protobuf:"bytes,2,opt,name=Committed,proto3" json:"Committed,omitempty"`
	Applied   *TailBitmap `protobuf:"bytes,3,opt,name=Applied,proto3" json:"Applied,omitempty"`
}
```


#### func (*ReplicaStatus) CmpAccepted

```go
func (a *ReplicaStatus) CmpAccepted(b *ReplicaStatus) int
```
CmpAccepted compares log related fields with another ballot. I.e. Committer and
MaxLogSeq.

#### func (*ReplicaStatus) Descriptor

```go
func (*ReplicaStatus) Descriptor() ([]byte, []int)
```

#### func (*ReplicaStatus) Equal

```go
func (this *ReplicaStatus) Equal(that interface{}) bool
```

#### func (*ReplicaStatus) GetAccepted

```go
func (m *ReplicaStatus) GetAccepted() *TailBitmap
```

#### func (*ReplicaStatus) GetApplied

```go
func (m *ReplicaStatus) GetApplied() *TailBitmap
```

#### func (*ReplicaStatus) GetCommitted

```go
func (m *ReplicaStatus) GetCommitted() *TailBitmap
```

#### func (*ReplicaStatus) GetCommitter

```go
func (m *ReplicaStatus) GetCommitter() *LeaderId
```

#### func (*ReplicaStatus) GetVotedFor

```go
func (m *ReplicaStatus) GetVotedFor() *LeaderId
```

#### func (*ReplicaStatus) Marshal

```go
func (m *ReplicaStatus) Marshal() (dAtA []byte, err error)
```

#### func (*ReplicaStatus) MarshalTo

```go
func (m *ReplicaStatus) MarshalTo(dAtA []byte) (int, error)
```

#### func (*ReplicaStatus) MarshalToSizedBuffer

```go
func (m *ReplicaStatus) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*ReplicaStatus) ProtoMessage

```go
func (*ReplicaStatus) ProtoMessage()
```

#### func (*ReplicaStatus) Reset

```go
func (m *ReplicaStatus) Reset()
```

#### func (*ReplicaStatus) Size

```go
func (m *ReplicaStatus) Size() (n int)
```

#### func (*ReplicaStatus) String

```go
func (m *ReplicaStatus) String() string
```

#### func (*ReplicaStatus) Unmarshal

```go
func (m *ReplicaStatus) Unmarshal(dAtA []byte) error
```

#### func (*ReplicaStatus) XXX_DiscardUnknown

```go
func (m *ReplicaStatus) XXX_DiscardUnknown()
```

#### func (*ReplicaStatus) XXX_Marshal

```go
func (m *ReplicaStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*ReplicaStatus) XXX_Merge

```go
func (m *ReplicaStatus) XXX_Merge(src proto.Message)
```

#### func (*ReplicaStatus) XXX_Size

```go
func (m *ReplicaStatus) XXX_Size() int
```

#### func (*ReplicaStatus) XXX_Unmarshal

```go
func (m *ReplicaStatus) XXX_Unmarshal(b []byte) error
```

#### type ReplicateReply

```go
type ReplicateReply struct {
}
```


#### func (*ReplicateReply) Descriptor

```go
func (*ReplicateReply) Descriptor() ([]byte, []int)
```

#### func (*ReplicateReply) Equal

```go
func (this *ReplicateReply) Equal(that interface{}) bool
```

#### func (*ReplicateReply) Marshal

```go
func (m *ReplicateReply) Marshal() (dAtA []byte, err error)
```

#### func (*ReplicateReply) MarshalTo

```go
func (m *ReplicateReply) MarshalTo(dAtA []byte) (int, error)
```

#### func (*ReplicateReply) MarshalToSizedBuffer

```go
func (m *ReplicateReply) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*ReplicateReply) ProtoMessage

```go
func (*ReplicateReply) ProtoMessage()
```

#### func (*ReplicateReply) Reset

```go
func (m *ReplicateReply) Reset()
```

#### func (*ReplicateReply) Size

```go
func (m *ReplicateReply) Size() (n int)
```

#### func (*ReplicateReply) String

```go
func (m *ReplicateReply) String() string
```

#### func (*ReplicateReply) Unmarshal

```go
func (m *ReplicateReply) Unmarshal(dAtA []byte) error
```

#### func (*ReplicateReply) XXX_DiscardUnknown

```go
func (m *ReplicateReply) XXX_DiscardUnknown()
```

#### func (*ReplicateReply) XXX_Marshal

```go
func (m *ReplicateReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*ReplicateReply) XXX_Merge

```go
func (m *ReplicateReply) XXX_Merge(src proto.Message)
```

#### func (*ReplicateReply) XXX_Size

```go
func (m *ReplicateReply) XXX_Size() int
```

#### func (*ReplicateReply) XXX_Unmarshal

```go
func (m *ReplicateReply) XXX_Unmarshal(b []byte) error
```

#### type TRaft

```go
type TRaft struct {
	Node
}
```


#### func (*TRaft) Replicate

```go
func (tr *TRaft) Replicate(ctx context.Context, req *Record) (*ReplicateReply, error)
```

#### func (*TRaft) Vote

```go
func (tr *TRaft) Vote(ctx context.Context, req *VoteReq) (*VoteReply, error)
```

#### type TRaftClient

```go
type TRaftClient interface {
	Vote(ctx context.Context, in *VoteReq, opts ...grpc.CallOption) (*VoteReply, error)
	Replicate(ctx context.Context, in *Record, opts ...grpc.CallOption) (*ReplicateReply, error)
}
```

TRaftClient is the client API for TRaft service.

For semantics around ctx use and closing/ending streaming RPCs, please refer to
https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.

#### func  NewTRaftClient

```go
func NewTRaftClient(cc *grpc.ClientConn) TRaftClient
```

#### type TRaftServer

```go
type TRaftServer interface {
	Vote(context.Context, *VoteReq) (*VoteReply, error)
	Replicate(context.Context, *Record) (*ReplicateReply, error)
}
```

TRaftServer is the server API for TRaft service.

#### type TailBitmap

```go
type TailBitmap struct {
	Offset   int64    `protobuf:"varint,1,opt,name=Offset,proto3" json:"Offset,omitempty"`
	Words    []uint64 `protobuf:"varint,2,rep,packed,name=Words,proto3" json:"Words,omitempty"`
	Reclamed int64    `protobuf:"varint,3,opt,name=Reclamed,proto3" json:"Reclamed,omitempty"`
}
```

TailBitmap is a bitmap that has all its leading bits set to `1`. Thus it is
compressed with an Offset of all-ones position and a trailing bitmap. It is used
to describe Record dependency etc.

The data structure is as the following described:

                       reclaimed
                       |
                       |     Offset
                       |     |
                       v     v
                 ..... X ... 01010...00111  00...
    bitIndex:    0123...     ^              ^
                             |              |
                             Words[0]       Words[1]

#### func  NewTailBitmap

```go
func NewTailBitmap(offset int64, set ...int64) *TailBitmap
```
NewTailBitmap creates an TailBitmap with a preset Offset and an empty tail
bitmap.

Optional arg `set` specifies what bit to set to 1. The bit positions in `set` is
absolute, NOT based on offset.

Since 0.1.22

#### func (*TailBitmap) Clone

```go
func (tb *TailBitmap) Clone() *TailBitmap
```

#### func (*TailBitmap) Compact

```go
func (tb *TailBitmap) Compact()
```
Compact all leading all-ones words in the bitmap.

Since 0.1.22

#### func (*TailBitmap) Descriptor

```go
func (*TailBitmap) Descriptor() ([]byte, []int)
```

#### func (*TailBitmap) Diff

```go
func (tb *TailBitmap) Diff(tc *TailBitmap)
```
Diff AKA substraction A - B or A \ B

#### func (*TailBitmap) Equal

```go
func (this *TailBitmap) Equal(that interface{}) bool
```

#### func (*TailBitmap) Get

```go
func (tb *TailBitmap) Get(idx int64) uint64
```
Get retrieves a bit at its 64-based offset.

Since 0.1.22

#### func (*TailBitmap) Get1

```go
func (tb *TailBitmap) Get1(idx int64) uint64
```
Get1 retrieves a bit and returns a 1-bit word, i.e., putting the bit in the
lowest bit.

Since 0.1.22

#### func (*TailBitmap) GetOffset

```go
func (m *TailBitmap) GetOffset() int64
```

#### func (*TailBitmap) GetReclamed

```go
func (m *TailBitmap) GetReclamed() int64
```

#### func (*TailBitmap) GetWords

```go
func (m *TailBitmap) GetWords() []uint64
```

#### func (*TailBitmap) Len

```go
func (tb *TailBitmap) Len() int64
```
Last returns last set bit index + 1.

#### func (*TailBitmap) Marshal

```go
func (m *TailBitmap) Marshal() (dAtA []byte, err error)
```

#### func (*TailBitmap) MarshalTo

```go
func (m *TailBitmap) MarshalTo(dAtA []byte) (int, error)
```

#### func (*TailBitmap) MarshalToSizedBuffer

```go
func (m *TailBitmap) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*TailBitmap) ProtoMessage

```go
func (*TailBitmap) ProtoMessage()
```

#### func (*TailBitmap) Reset

```go
func (m *TailBitmap) Reset()
```

#### func (*TailBitmap) Set

```go
func (tb *TailBitmap) Set(idx int64)
```
Set the bit at `idx` to `1`.

Since 0.1.22

#### func (*TailBitmap) Size

```go
func (m *TailBitmap) Size() (n int)
```

#### func (*TailBitmap) String

```go
func (m *TailBitmap) String() string
```

#### func (*TailBitmap) Union

```go
func (tb *TailBitmap) Union(tc *TailBitmap)
```

#### func (*TailBitmap) Unmarshal

```go
func (m *TailBitmap) Unmarshal(dAtA []byte) error
```

#### func (*TailBitmap) XXX_DiscardUnknown

```go
func (m *TailBitmap) XXX_DiscardUnknown()
```

#### func (*TailBitmap) XXX_Marshal

```go
func (m *TailBitmap) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*TailBitmap) XXX_Merge

```go
func (m *TailBitmap) XXX_Merge(src proto.Message)
```

#### func (*TailBitmap) XXX_Size

```go
func (m *TailBitmap) XXX_Size() int
```

#### func (*TailBitmap) XXX_Unmarshal

```go
func (m *TailBitmap) XXX_Unmarshal(b []byte) error
```

#### type UnimplementedTRaftServer

```go
type UnimplementedTRaftServer struct {
}
```

UnimplementedTRaftServer can be embedded to have forward compatible
implementations.

#### func (*UnimplementedTRaftServer) Replicate

```go
func (*UnimplementedTRaftServer) Replicate(ctx context.Context, req *Record) (*ReplicateReply, error)
```

#### func (*UnimplementedTRaftServer) Vote

```go
func (*UnimplementedTRaftServer) Vote(ctx context.Context, req *VoteReq) (*VoteReply, error)
```

#### type VoteReply

```go
type VoteReply struct {
	// voted for a candidate or the previous voted other leader.
	VotedFor *LeaderId `protobuf:"bytes,10,opt,name=VotedFor,proto3" json:"VotedFor,omitempty"`
	// latest log committer.
	Committer *LeaderId `protobuf:"bytes,4,opt,name=Committer,proto3" json:"Committer,omitempty"`
	// what logs I have.
	Accepted *TailBitmap `protobuf:"bytes,1,opt,name=Accepted,proto3" json:"Accepted,omitempty"`
	// The logs that voter has but leader candidate does not have.
	// For the leader to rebuild all possibly committed logs from a quorum.
	Logs []*Record `protobuf:"bytes,30,rep,name=Logs,proto3" json:"Logs,omitempty"`
}
```


#### func (*VoteReply) Descriptor

```go
func (*VoteReply) Descriptor() ([]byte, []int)
```

#### func (*VoteReply) Equal

```go
func (this *VoteReply) Equal(that interface{}) bool
```

#### func (*VoteReply) GetAccepted

```go
func (m *VoteReply) GetAccepted() *TailBitmap
```

#### func (*VoteReply) GetCommitter

```go
func (m *VoteReply) GetCommitter() *LeaderId
```

#### func (*VoteReply) GetLogs

```go
func (m *VoteReply) GetLogs() []*Record
```

#### func (*VoteReply) GetVotedFor

```go
func (m *VoteReply) GetVotedFor() *LeaderId
```

#### func (*VoteReply) Marshal

```go
func (m *VoteReply) Marshal() (dAtA []byte, err error)
```

#### func (*VoteReply) MarshalTo

```go
func (m *VoteReply) MarshalTo(dAtA []byte) (int, error)
```

#### func (*VoteReply) MarshalToSizedBuffer

```go
func (m *VoteReply) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*VoteReply) ProtoMessage

```go
func (*VoteReply) ProtoMessage()
```

#### func (*VoteReply) Reset

```go
func (m *VoteReply) Reset()
```

#### func (*VoteReply) Size

```go
func (m *VoteReply) Size() (n int)
```

#### func (*VoteReply) String

```go
func (m *VoteReply) String() string
```

#### func (*VoteReply) Unmarshal

```go
func (m *VoteReply) Unmarshal(dAtA []byte) error
```

#### func (*VoteReply) XXX_DiscardUnknown

```go
func (m *VoteReply) XXX_DiscardUnknown()
```

#### func (*VoteReply) XXX_Marshal

```go
func (m *VoteReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*VoteReply) XXX_Merge

```go
func (m *VoteReply) XXX_Merge(src proto.Message)
```

#### func (*VoteReply) XXX_Size

```go
func (m *VoteReply) XXX_Size() int
```

#### func (*VoteReply) XXX_Unmarshal

```go
func (m *VoteReply) XXX_Unmarshal(b []byte) error
```

#### type VoteReq

```go
type VoteReq struct {
	// who initiates the election
	Candidate *LeaderId `protobuf:"bytes,1,opt,name=Candidate,proto3" json:"Candidate,omitempty"`
	// Latest leader that forwarded log to the candidate
	Committer *LeaderId `protobuf:"bytes,2,opt,name=Committer,proto3" json:"Committer,omitempty"`
	// what logs the candidate has.
	Accepted *TailBitmap `protobuf:"bytes,3,opt,name=Accepted,proto3" json:"Accepted,omitempty"`
}
```


#### func (*VoteReq) Descriptor

```go
func (*VoteReq) Descriptor() ([]byte, []int)
```

#### func (*VoteReq) Equal

```go
func (this *VoteReq) Equal(that interface{}) bool
```

#### func (*VoteReq) GetAccepted

```go
func (m *VoteReq) GetAccepted() *TailBitmap
```

#### func (*VoteReq) GetCandidate

```go
func (m *VoteReq) GetCandidate() *LeaderId
```

#### func (*VoteReq) GetCommitter

```go
func (m *VoteReq) GetCommitter() *LeaderId
```

#### func (*VoteReq) Marshal

```go
func (m *VoteReq) Marshal() (dAtA []byte, err error)
```

#### func (*VoteReq) MarshalTo

```go
func (m *VoteReq) MarshalTo(dAtA []byte) (int, error)
```

#### func (*VoteReq) MarshalToSizedBuffer

```go
func (m *VoteReq) MarshalToSizedBuffer(dAtA []byte) (int, error)
```

#### func (*VoteReq) ProtoMessage

```go
func (*VoteReq) ProtoMessage()
```

#### func (*VoteReq) Reset

```go
func (m *VoteReq) Reset()
```

#### func (*VoteReq) Size

```go
func (m *VoteReq) Size() (n int)
```

#### func (*VoteReq) String

```go
func (m *VoteReq) String() string
```

#### func (*VoteReq) Unmarshal

```go
func (m *VoteReq) Unmarshal(dAtA []byte) error
```

#### func (*VoteReq) XXX_DiscardUnknown

```go
func (m *VoteReq) XXX_DiscardUnknown()
```

#### func (*VoteReq) XXX_Marshal

```go
func (m *VoteReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*VoteReq) XXX_Merge

```go
func (m *VoteReq) XXX_Merge(src proto.Message)
```

#### func (*VoteReq) XXX_Size

```go
func (m *VoteReq) XXX_Size() int
```

#### func (*VoteReq) XXX_Unmarshal

```go
func (m *VoteReq) XXX_Unmarshal(b []byte) error
```
