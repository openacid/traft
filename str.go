package traft

import "fmt"

type shortStrer interface {
	ShortStr() string
}

func toStr(v interface{}) string {

	switch ss := v.(type) {
	case shortStrer:
		return ss.ShortStr()
	case fmt.Stringer:
		return ss.String()
	default:
		return fmt.Sprintf("%v", v)

	}
}
