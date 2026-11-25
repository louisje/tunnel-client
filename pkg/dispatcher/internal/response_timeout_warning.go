package dispatcherinternal

import (
	"context"
	"time"
)

const responseTimeoutWarningLead = 10 * time.Second

func waitForResponseTimeoutWarning(ctx context.Context, deadline time.Time) bool {
	timer := time.NewTimer(max(0, time.Until(deadline)-responseTimeoutWarningLead))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return ctx.Err() == nil && time.Now().Before(deadline)
	}
}
