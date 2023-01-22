package middleware

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

func NewDefaultTaskLogMiddleware(logger log.Logger) processor.MiddlewareFunc {
	return func(next processor.ProcessFunc) processor.ProcessFunc {
		return func(ctx context.Context) error {
			logger.ForCtx(ctx).Info("task started")
			defer logger.ForCtx(ctx).Info("task finished")

			return next(ctx)
		}
	}
}
