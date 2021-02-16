package traft

import "github.com/pkg/errors"

var (
	ErrStaleLog    = errors.New("local log is stale")
	ErrStaleTermId = errors.New("local Term-Id is stale")
	ErrTimeout     = errors.New("timeout")
	ErrLeaderLost  = errors.New("leadership lost")
)
