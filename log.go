package traft

var (
	emptyRecord = &Record{}
)

// a func for test purpose only
func (tr *TRaft) addLogs(cmds ...interface{}) {
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

func (tr *TRaft) GetLog(lsn int64) *Record {
	idx := lsn - tr.LogOffset
	r := tr.Logs[idx]
	if r.Seq != lsn {
		panic("wtf")
	}
	return r
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
