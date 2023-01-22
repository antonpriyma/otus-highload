package middleware

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/debug"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

func LogErrorHandler(mainLogger log.Logger) processor.ErrorHandler {
	return func(ctx context.Context, err error) {
		logger := mainLogger.ForCtx(ctx).WithError(err)

		if errors.Is(err, errors.ErrPanic) {
			logger = logger.WithField("stack", debug.ErrorStackTrace(err))
		}

		logFunc := logger.Error
		if errors.Is(err, processor.ErrDeleteTask) {
			logFunc = logger.Warn
		}
		if errors.Is(err, processor.ErrRetryTask) {
			logFunc = logger.Info
		}

		logFunc("failed to process task")
	}
}
