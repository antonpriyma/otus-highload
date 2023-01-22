package retrier

import (
	"context"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

type Config struct {
	Retries int           `mapstructure:"retries"`
	Sleep   time.Duration `mapstructure:"sleep"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type Retrier struct {
	Config
	IsUnrecoverableError func(error) bool
	Logger               log.Logger
}

func (r Retrier) Do(ctx context.Context, call func() error) (err error) {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	var ticker *time.Ticker
	if r.Sleep > 0 {
		ticker = time.NewTicker(r.Sleep)
		defer ticker.Stop()
	}

	for i := 0; ; i++ {
		err = call()
		if err == nil || r.isUnrecoverable(err) {
			return err
		}

		if r.Logger != nil {
			r.Logger.ForCtx(ctx).WithError(err).Warnf("attempt #%d failed", i+1)
		}

		if r.Retries > 0 && i >= r.Retries {
			return err
		}

		// we need to check context error because call() can do no context check inside itself
		// while it takes time
		// so, we should not sleep in that case and should break cycle
		if ctx.Err() != nil {
			return errors.Wrap(err, "retries canceled by context")
		}

		if ticker != nil {
			select {
			case <-ctx.Done():
				return errors.Wrap(err, "retries canceled by context")
			case <-ticker.C:
			}
		}
	}
}

func (r Retrier) isUnrecoverable(err error) bool {
	return r.IsUnrecoverableError != nil && r.IsUnrecoverableError(err)
}

func SleepOrDone(ctx context.Context, t time.Duration) error {
	sleepTimer := time.NewTimer(t)
	defer sleepTimer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sleepTimer.C:
		return nil
	}
}
