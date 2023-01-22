package middleware

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/context/reqid"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

func NewRequestIDMiddleware(reqIDGetter func(context.Context) string) processor.MiddlewareFunc {
	return func(next processor.ProcessFunc) processor.ProcessFunc {
		return func(ctx context.Context) (err error) {
			// passed from api
			reqID := reqIDGetter(ctx)

			ctx = reqid.SetRequestID(ctx, reqID)
			ctx = log.AddCtxFields(ctx, log.Fields{
				"request_id": reqID,
			})

			return next(ctx)
		}
	}
}
