package middleware

import (
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat/loggerstat"
	"github.com/labstack/echo"
)

func LoggerStatMiddleware(logger log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := loggerstat.InitStatForCtx(c.Request().Context())
			c.SetRequest(c.Request().WithContext(ctx))
			defer loggerstat.PrintStat(ctx, logger)
			return next(c)
		}
	}
}
