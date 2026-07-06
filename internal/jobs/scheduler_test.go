package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeriodicTasks_IncludesMandatePoll(t *testing.T) {
	found := false
	for _, e := range periodicTaskEntries() {
		if e.task == TaskMandatePollStatus {
			found = true
			require.Equal(t, "*/15 * * * *", e.cron)
			break
		}
	}
	require.True(t, found, "mandate poll task should be scheduled")
}
