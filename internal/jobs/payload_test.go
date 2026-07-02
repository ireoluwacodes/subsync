package jobs

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParseIDs(t *testing.T) {
	tid := uuid.New()
	sid := uuid.New()
	gotTenant, gotSub, err := parseIDs(tid.String(), sid.String())
	require.NoError(t, err)
	require.Equal(t, tid, gotTenant)
	require.Equal(t, sid, gotSub)
}
