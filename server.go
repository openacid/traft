package traft

// TRaftServer impl

import (
	"time"
)

var leaderLease = int64(time.Second * 1)

// init a TRaft for test, all logs are `set x=lsn`
func (tr *TRaft) initTraft0(committer *LeaderId, votedFor *LeaderId, cmds...interface{}) {
	id := tr.Id

	tr.Status[id].Committer = committer.Clone()
	tr.Status[id].VotedFor = votedFor.Clone()

	tr.addlogs(cmds...)

	tr.checkStatus()
}

// init a TRaft for test, all logs are `set x=lsn`
func (tr *TRaft) initTraft(
	// proposer of the logs
	committer *LeaderId,
	// author of the logs
	author *LeaderId,
	// log seq numbers to generate.
	lsns []int64,
	nilLogs map[int64]bool,
	committed []int64,
	votedFor *LeaderId,
) {
	id := tr.Id

	tr.LogOffset, tr.Logs = buildPseudoLogs(author, lsns, nilLogs)

	tr.Status[id].Committer = committer.Clone()
	tr.Status[id].Accepted = NewTailBitmap(0, lsns...)

	if committed == nil {
		tr.Status[id].Committed = NewTailBitmap(0)
	} else {
		tr.Status[id].Committed = NewTailBitmap(0, committed...)
	}

	tr.Status[id].VotedFor = votedFor.Clone()

	tr.checkStatus()
}

func buildPseudoLogs(
	// author of the logs
	author *LeaderId,
	// log seq numbers to generate.
	lsns []int64,
	nilLogs map[int64]bool,
) (int64, []*Record) {
	logs := make([]*Record, 0)
	if len(lsns) == 0 {
		return 0, logs
	}

	last := lsns[len(lsns)-1]
	start := lsns[0]
	for i := start; i <= last; i++ {
		logs = append(logs, &Record{})
	}

	for _, lsn := range lsns {
		if nilLogs != nil && nilLogs[lsn] {
		} else {
			logs[lsn-start] = NewRecord(
				author.Clone(),
				lsn,
				NewCmdI64("set", "x", lsn))
		}
	}
	return start, logs
}

// a func for test purpose only
func (tr *TRaft) addlogs(cmds ...interface{}) {
	me := tr.Status[tr.Id]
	for _, cs := range cmds {
		cmd := toCmd(cs)
		r := tr.addLogInternal(cmd)
		me.Accepted.Union(r.Overrides)
	}
}

// Only a established leader should use this func.
// no lock protection, must be called from Loop()
func (tr *TRaft) AddLog(cmd *Cmd) *Record {

	me := tr.Status[tr.Id]

	if me.VotedFor.Id != tr.Id {
		panic("wtf")
	}

	return tr.addLogInternal(cmd)
}

func (tr *TRaft) addLogInternal(cmd *Cmd) *Record {

	me := tr.Status[tr.Id]

	lsn := tr.LogOffset + int64(len(tr.Logs))

	r := NewRecord(me.VotedFor.Clone(), lsn, cmd)

	// find the first interfering record.

	var i int
	for i = len(tr.Logs) - 1; i >= 0; i-- {
		prev := tr.Logs[i]
		if r.Interfering(prev) {
			r.Overrides = prev.Overrides.Clone()
			break
		}
	}

	if i == -1 {
		// there is not a interfering record.
		r.Overrides = NewTailBitmap(0)
	}

	r.Overrides.Set(lsn)

	// all log I do not know must be executed in order.
	// Because I do not know of the intefering relations.
	r.Depends = NewTailBitmap(tr.LogOffset)

	// reduce bitmap size by removing unknown logs
	r.Overrides.Union(NewTailBitmap(tr.LogOffset & ^63))

	tr.Logs = append(tr.Logs, r)

	return r
}
