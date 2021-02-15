package traft

import (
	fmt "fmt"
	"strconv"
	"strings"
)

func NewCmdI64(op, key string, v int64) *Cmd {
	cmd := &Cmd{
		Op:    op,
		Key:   key,
		Value: &Cmd_VI64{v},
	}
	return cmd
}

func cmdValueShortStr(v isCmd_Value) string {
	switch vv := v.(type) {
	case *Cmd_VI64:
		return fmt.Sprintf("%d", vv.VI64)
	case *Cmd_VStr:
		return vv.VStr
		// TODO ClusterConfig
	default:
		return fmt.Sprintf("%s", vv)
	}
}

func (c *Cmd) ShortStr() string {
	if c == nil {
		return "()"
	}
	return fmt.Sprintf("%s(%s, %s)",
		c.Op, c.Key, cmdValueShortStr(c.Value))
}

// Interfering check if a command interferes with another one,
// i.e. they change the same key.
func (a *Cmd) Interfering(b *Cmd) bool {
	if a == nil || b == nil {
		return false
	}

	if a.Op == "set" && b.Op == "set" {
		if a.Key == b.Key {
			return true
		}
	}

	return false
}

type toCmder interface {
	ToCmd() *Cmd
}

type cstr string

func (c *cstr) ToCmd() *Cmd {
	if *c == "" {
		return nil
	}

	kv := strings.Split(string(*c), "=")
	k := kv[0]

	v, err := strconv.ParseInt(kv[1], 10, 64)
	if err != nil {
		panic(string(*c) + " convert to Cmd")
	}
	return NewCmdI64("set", k, v)
}

func toCmd(x interface{}) *Cmd {
	if x == nil {
		return nil
	}

	switch v := x.(type) {
	case string:
		s := cstr(v)
		return s.ToCmd()

	case *Cmd:
		return v
	}
	panic("invalid type to convert to cmd")
}
