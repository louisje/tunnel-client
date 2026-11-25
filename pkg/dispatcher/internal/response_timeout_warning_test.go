package dispatcherinternal

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestResponseTimeoutWarningTimer(t *testing.T) {
	deadline := time.Now().Add(responseTimeoutWarningLead + 50*time.Millisecond)
	require.True(t, waitForResponseTimeoutWarning(context.Background(), deadline))
	require.LessOrEqual(t, time.Until(deadline), responseTimeoutWarningLead)
	require.Greater(t, time.Until(deadline), 9*time.Second)
}
func TestResponseTimeoutWarningCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.False(t, waitForResponseTimeoutWarning(ctx, time.Now().Add(time.Minute)))
	require.False(t, waitForResponseTimeoutWarning(context.Background(), time.Now().Add(-time.Second)))
}
