package traft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTRaft_Vote(t *testing.T) {

	ta := require.New(t)

	ids := []int64{1, 2, 3}

	servers, trafts := serveCluster(ids)

	defer func() {
		for _, s := range servers {
			s.Stop()
		}
	}()

	tr := trafts[0]
	tr.LogOffset = 5
	tr.Log = append(tr.Log)

	ta.True(true)

}
