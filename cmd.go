package traft

import (
	fmt "fmt"
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
	// return c.String()
	// return proto.CompactTextString(c.Value)
	return fmt.Sprintf("%s(%s, %s)",
		c.Op, c.Key, cmdValueShortStr(c.Value))
}
