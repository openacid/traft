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

func NewRecordOverride(leader *LeaderId, seq int64, cmd *Cmd, override *TailBitmap) *Record {

	rec := NewRecord(leader,seq,cmd)
	rec.Overrides = NewTailBitmap(0, seq)
	rec.Overrides.Union(override)

	return rec
}

// gogoproto would panic if a []*Record has a nil in it.
// Thus we use r.Cmd == nil  to indicate an absent log record.
func (r *Record) Empty() bool {
	return r == nil || r.Cmd == nil
}

func (a *Record) Interfering(b *Record) bool {
	if a == nil || b == nil {
		return false
	}

	return a.Cmd.Interfering(b.Cmd)
}

func (r *Record) ShortStr() string {
	if r.Empty() {
		return "<>"
	}

	return fmt.Sprintf("<%s:%03d{%s}-%s→%s>",
		r.Author.ShortStr(),
		r.Seq,
		r.Cmd.ShortStr(),
		r.Overrides.ShortStr(),
		r.Depends.ShortStr(),
	)
}

func RecordsShortStr(rs []*Record, sep ...string) string {
	s := ", "
	if len(sep) > 0 {
		s = sep[0]
	}
	rst := []string{}
	for _, r := range rs {
		rst = append(rst, r.ShortStr())
	}
	return "[" + strings.Join(rst, s) + "]"

}
