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

	tr.addLogs(cmds...)

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
) (int64, []*LogRecord) {
	logs := make([]*LogRecord, 0)
	if len(lsns) == 0 {
		return 0, logs
	}

	last := lsns[len(lsns)-1]
	start := lsns[0]
	for i := start; i <= last; i++ {
		logs = append(logs, &LogRecord{})
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

