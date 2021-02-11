package traft

func NewCmdI64(op, key string, v int64) *Cmd {
	cmd := &Cmd{
		Op:    op,
		Key:   key,
		Value: &Cmd_VI64{v},
	}
	return cmd
}
