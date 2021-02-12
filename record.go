package traft

import (
	fmt "fmt"
	"strings"
)

// NewRecord: without Overrides yet!!! TODO
func NewRecord(leader *LeaderId, seq int64, cmd *Cmd) *Record {

	rec := &Record{
		Author: leader,
		Seq:    seq,
		Cmd:    cmd,
	}

	return rec
}

// gogoproto would panic if a []*Record has a nil in it.
// Thus we use r.Cmd == nil  to indicate an absent log record.
func (r *Record) Empty() bool {
	return r == nil || r.Cmd == nil
}

func (r *Record) ShortStr() string {
	if r.Empty() {
		return "<>"
	}

	return fmt.Sprintf("<%s:%03d{%s}>",
		r.Author.ShortStr(),
		r.Seq,
		r.Cmd.ShortStr())
}

func RecordsShortStr(rs []*Record) string {
	rst := []string{}
	for _, r := range rs {
		rst = append(rst, r.ShortStr())
	}
	return "[" + strings.Join(rst, ", ") + "]"

}
