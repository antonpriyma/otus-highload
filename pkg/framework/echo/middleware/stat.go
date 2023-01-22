package middleware

import (
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/labstack/echo"
)

func StatAroundResponse(registry stat.Registry, routeFilter *echoutils.RouteFilter) BeforeFunc {
	var statSender struct {
		RequestDuration stat.TimerCtor `labels:"api_status"`
	}
	stat.NewRegistrar(registry.ForSubsystem("middleware")).MustRegister(&statSender)

	return func(c echo.Context) AfterFunc {
		timer := statSender.RequestDuration.Timer(c.Request().Context()).WithLabels(stat.Labels{
			"api_end":    routeFilter.GetFilteredPath(c),
			"api_method": c.Request().Method,
		}).Start()

		return func(c echo.Context) {
			timer.WithLabels(stat.Labels{"api_status": echoutils.GetResponseStatus(c)}).Stop()
		}
	}
}
