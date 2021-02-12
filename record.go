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

func (r *Record) ShortStr() string {
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
