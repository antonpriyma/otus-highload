package middleware

import (
	"github.com/antonpriyma/otus-highload/pkg/context/request"
	"github.com/labstack/echo"
)

func SetCtxReqMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()

		ctx := c.Request().Context()
		ctx = request.SetRequest(ctx, req)

		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
