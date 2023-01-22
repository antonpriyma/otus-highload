package middleware

import (
	"context"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

type RatelimitConfig struct {
	Period time.Duration `mapstructure:"period"`
}

func NewRatelimitMiddleware(cfg RatelimitConfig, logger log.Logger) processor.MiddlewareFunc {
	if cfg.Period == 0 {
		return emptyMiddleware
	}

	ticker := time.NewTicker(cfg.Period)

	return func(next processor.ProcessFunc) processor.ProcessFunc {
		return func(ctx context.Context) (err error) {
			select {
			case <-ticker.C:
				logger.ForCtx(ctx).Info("ratelimit passed, continue")
				return next(ctx)
			case <-ctx.Done():
				return errors.Wrap(ctx.Err(), "ratelimit interrupted by context")
			}
		}
	}
}

func emptyMiddleware(next processor.ProcessFunc) processor.ProcessFunc {
	return next
}
