package traft

// NewRecord: without Overrides yet!!! TODO
func NewRecord(leader *LeaderId, seq int64, cmd *Cmd) *Record {

	rec := &Record{
		By:  leader,
		Seq: seq,
		Cmd: cmd,
	}

	return rec
}
