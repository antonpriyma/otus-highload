package middleware

import (
	"github.com/antonpriyma/otus-highload/pkg/clients/sentry"
	"github.com/labstack/echo"
)

func SentryMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		ctx = sentry.InitContextExtra(ctx)

		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
