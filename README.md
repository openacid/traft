<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [traft](#traft)
- [Features:](#features)
- [Day-0 2021-02-09](#day-0-2021-02-09)
- [Day-1 2021-02-10](#day-1-2021-02-10)
- [Day-2 2021-02-11](#day-2-2021-02-11)
- [Day-3 2021-02-12](#day-3-2021-02-12)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# traft

都说deadline 是第一生产力. 放假没事, 试试看能不能7天写个raft出来, 当练手了. 2021-02-09
装逼有风险, 万一写不出来也别笑话我.

# Features:

- [ ] Leader election
- [ ] WAL log
- [ ] snapshot: impl with https://github.com/openacid/slim , a static kv-like storage engine supporting protobuf.
- [ ] member change with generalized joint consensus.
- [ ] Out of order commit/apply if possible.

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