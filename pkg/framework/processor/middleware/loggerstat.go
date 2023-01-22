package middleware

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat/loggerstat"
)

func NewLoggerStatMiddleware(logger log.Logger) processor.MiddlewareFunc {
	return func(next processor.ProcessFunc) processor.ProcessFunc {
		return func(ctx context.Context) (err error) {
			ctx = loggerstat.InitStatForCtx(ctx)
			defer loggerstat.PrintStat(ctx, logger)

			return next(ctx)
		}
	}
}
