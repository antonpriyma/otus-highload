package middleware

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

func NewRecoverMiddleware(logger log.Logger) processor.MiddlewareFunc {
	return func(next processor.ProcessFunc) processor.ProcessFunc {
		return func(ctx context.Context) (err error) {
			defer func() {
				r := errors.RecoverError(recover())
				if r != nil {
					err = r
				}
			}()

			return next(ctx)
		}
	}
}
