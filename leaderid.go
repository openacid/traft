package traft

import (
	fmt "fmt"

	proto "github.com/gogo/protobuf/proto"
)

func NewLeaderId(term, id int64) *LeaderId {
	return &LeaderId{
		Term: term,
		Id:   id,
	}
}

// Compare two leader id and returns 1, 0 or -1 for greater, equal and less
func (a *LeaderId) Cmp(b *LeaderId) int {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}

	if b == nil {
		return 1
	}

	r := cmpI64(a.Term, b.Term)
	if r != 0 {
		return r
	}

	return cmpI64(a.Id, b.Id)
}

func (l *LeaderId) Clone() *LeaderId {
	return proto.Clone(l).(*LeaderId)
}

func (l *LeaderId) ShortStr() string {
	return fmt.Sprintf("%03d#%03d", l.Term, l.Id)
}
