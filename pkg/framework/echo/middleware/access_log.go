package middleware

import (
	"time"

	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/labstack/echo"
)

func AccessLogAfterware(logger log.Logger, routeFilter *echoutils.RouteFilter) BeforeFunc {
	return func(c echo.Context) AfterFunc {
		start := time.Now()

		ctx := c.Request().Context()
		ctx = log.AddCtxFields(ctx, log.Fields{
			"method": c.Request().Method,
			"path":   routeFilter.GetFilteredPath(c),
		})
		c.SetRequest(c.Request().WithContext(ctx))

		return func(context echo.Context) {
			header := c.Request().Header
			logger.ForCtx(ctx).WithFields(log.Fields{
				"user_agent":    header.Get(echoutils.HeaderUserAgent),
				"referer":       header.Get(echoutils.HeaderReferer),
				"user_ip":       header.Get(echoutils.HeaderUserIP),
				"original_host": header.Get(echoutils.HeaderOriginalHost),
				"resolved_host": header.Get(echoutils.HeaderResolvedHost),
				"front":         header.Get(echoutils.HeaderFront),
				"request_url":   c.Request().URL.String(),
				"duration":      time.Since(start).String(),
				"status":        echoutils.GetResponseStatus(c),
			}).Info("access log")
		}
	}
}
