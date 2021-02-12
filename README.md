<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [traft](#traft)
- [Planned improvement to original raft](#planned-improvement-to-original-raft)
- [Features:](#features)
- [Progress](#progress)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# traft

都说deadline 是第一生产力. 放假没事, 试试看能不能7天写个raft出来, 当练手了. 2021-02-09
装逼有风险, 万一写不出来也别笑话我.

7天内直播: https://live.bilibili.com/22732677?visit_id=6eoy1n42a1w0

# Planned improvement to original raft

-   Leader election with less conflict.
    Raft has the issue that in a term, if multiple candidates try to elect theirselves,
    the conflicts are unacceptable.
    Because in one term, A voter will only vote for at most one leader.

    In our impl, `term` is less strict:
    A voter is allowed to vote another GREATER leader.

-   Out of order commit out of order apply.

    We use a bitmap to describe what logs interfere each other.

    
-   Adopt a more generalized member change algo.

    Get rid of single-step member change.
    Because any of the CORRECT single-step member change algo is equivalent to joint-consensus.

    But joint-consensus is a specialization of the TODO (I dont have good name for it yet).



# Features:

- [ ] Leader election
- [ ] WAL log
- [ ] snapshot: impl with https://github.com/openacid/slim , a static kv-like storage engine supporting protobuf.
- [ ] member change with generalized joint consensus.
- [ ] Out of order commit/apply if possible.

# Progress

- day-1: buildMajorityQuorums()
- day-1: import TailBitmap, add Union()
- day-1: TailBitmap.Clone()
- day-1: TailBitmap.Diff()
- day-1: refactor buildMajorityQuorums()
- day-2: LeaderId.Cmp()
- day-2: add NewLeaderId
- day-2: add NewBallot()
- day-2: fix: LeaderId.Cmp() accept nil as operand
- day-2: Ballot.CmpLog() to compare only the log-related fields
- day-2: rename Ballot.Accepted to Ballot.AcceptedFrom
- day-2: TailBitmap.Len()
- day-2: refactor design, add ReplicaStatus, remove Ballot.
- day-2: impl Vote Handler, under dev!!!
- day-2: use gogoproto, build clean .pb.go with less code, add serveCluster() to setup a simple cluster for test
- day-2: add NewCmdI64()
- day-2: add NewRecord()
- day-2: add ClusterConfig.GetReplicaInfo()
- day-2: NewTailBitmap() accepts extra bits to set
- day-2: add LeaderId:Clone()
- day-2: refactor TailBitmap.Clone(): use proto
- day-3: draft test of vote
- day-3: rename By to Author, AcceptedFrom to Committer. borrowed concepts from git:DDD
- day-3: add test: HandleVoteReq. add ShortStr() to Cmd, LeaderI, Record and []Record
- day-3: refactor VoteReply: do not use ReplicaStatus to describe log status
- day-3: add readme to record progress
- day-3: readme: impl storage with slim
- day-3: test that voter send back nil log
- day-3: update readme: collect git log
- day-3: TailBitmap: accept second operand to be nil
- day-3: test Replicate, under dev
- day-3: rename Node.Log to Node.Logs
- day-3: add Cmd.Intefering() to check if two commands not be allowed to change execution order
- day-3: add Record.Interfering()



<!--
# Day-0 2021-02-09

- [x] TailBitmap to support for describing log dependency etc. see: https://github.com/openacid/low/blob/ci/bitmap/tailbitmap.go

# Day-1 2021-02-10

- [x]: design t-raft protobuf

# Day-2 2021-02-11

- [x]: design t-raft protobuf
- [x]: impl handle_vote

# Day-3 2021-02-12

- [x]: refactor concepts
- [x]: test handle_vote
- [ ]: impl log replication
- [ ]: impl traft main-loop
-->