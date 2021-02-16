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

-   2021-02-10:
    LOC: +1008 -1

    ```
    buildMajorityQuorums()
    import TailBitmap, add Union()
    TailBitmap.Clone()
    TailBitmap.Diff()
    ```

-   2021-02-11:
    LOC: +754 -44

    ```
    refactor buildMajorityQuorums()
    LeaderId.Cmp()
    add NewLeaderId
    add NewBallot()
    fix: LeaderId.Cmp() accept nil as operand
    Ballot.CmpLog() to compare only the log-related fields
    rename Ballot.Accepted to Ballot.AcceptedFrom
    TailBitmap.Len()
    refactor design, add ReplicaStatus, remove Ballot.
    impl Vote Handler, under dev!!!
    use gogoproto, build clean .pb.go with less code, add serveCluster() to setup a simple cluster for test
    add NewCmdI64()
    add NewRecord()
    add ClusterConfig.GetReplicaInfo()
    NewTailBitmap() accepts extra bits to set
    ```

-   2021-02-12:
    LOC: +2601 -70

    ```
    add LeaderId:Clone()
    refactor TailBitmap.Clone(): use proto
    draft test of vote
    rename By to Author, AcceptedFrom to Committer. borrowed concepts from git:DDD
    add test: HandleVoteReq. add ShortStr() to Cmd, LeaderI, Record and []Record
    refactor VoteReply: do not use ReplicaStatus to describe log status
    add readme to record progress
    readme: impl storage with slim
    test that voter send back nil log
    update readme: collect git log
    TailBitmap: accept second operand to be nil
    ```

-   2021-02-13:
    LOC: +1309 -183

    ```
    test Replicate, under dev
    rename Node.Log to Node.Logs
    add Cmd.Intefering() to check if two commands not be allowed to change execution order
    add Record.Interfering()
    update readme
    NewTailBitmap() support non-64-aligned offset
    add AddLog() for leader to propose a command
    use map to store cluster members instead of slice
    add ClusterConfig.IsQuorum() to check if a set of members is a quorum
    TRaft.VoteOnce() run one round voting, to establish leadership
    TRaft actor loop and voting loop
    ```

-   2021-02-14:
    LOC: +891 -229

    ```
    mainloop as an actor is the only one update traft data. under dev. not passed yet
    test VoteLoop
    granted leader merges collected logs
    after vote, leader merge responded logs
    add API Propose without replication
    ```

-   2021-02-15:
    LOC: +476 -197

    ```
    add interface toCmd() to build Cmd from string
    test Replicate()
    add TRaft.sleep(): sleep only if it is not stopped.
    ```

-   2021-02-16:
    LOC: +1297 -592

    ```
    refactor: rename Replicate to LogForward
    refactor: vote_test: remove unused type, funcs
    gruop daily log format by date
    ```




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