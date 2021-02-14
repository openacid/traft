package traft

// if the first log in v.Logs matches lsn, pop and return it.
// Otherwise return nil.
func (v *VoteReply) PopRecord(lsn int64) *Record {
	if len(v.Logs) == 0 {
		return nil
	}

	r := v.Logs[0]
	if r.Seq < lsn {
		panic("wtf")
	}

	if r.Seq == lsn {
		v.Logs = v.Logs[1:]
		return r
	}

	return nil

}
