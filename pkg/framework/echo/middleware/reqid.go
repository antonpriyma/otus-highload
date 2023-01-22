package middleware

import (
	"fmt"
	"math/rand"

	"github.com/antonpriyma/otus-highload/pkg/context/reqid"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/labstack/echo"
)

func SetCtxReqIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()

		requestID := req.Header.Get(echoutils.HeaderRequestID)
		if requestID == "" {
			requestID = fmt.Sprintf("%016x", rand.Int()) // nolint: gosec
		}

		c.Response().Header().Set(echoutils.HeaderRequestID, requestID)

		ctx := c.Request().Context()
		ctx = reqid.SetRequestID(ctx, requestID)
		ctx = log.AddCtxFields(ctx, log.Fields{"request_id": requestID})

		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
