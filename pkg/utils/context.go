package utils

import (
	"context"
	"fmt"
	"time"
)

func SleepOrDone(ctx context.Context, t time.Duration) error {
	if t == 0 {
		return nil
	}

	sleepTimer := time.NewTimer(t)
	defer sleepTimer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sleepTimer.C:
		return nil
	}
}

// CheckDone checks if context is terminated
func CheckDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("terminating: %s", ctx.Err())
	default:
		return nil
	}
}
