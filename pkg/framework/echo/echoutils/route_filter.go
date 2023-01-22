package echoutils

import (
	"sync"

	"github.com/labstack/echo"
)

func NewRouteFilter(e *echo.Echo) *RouteFilter {
	return &RouteFilter{
		Echo: e,
	}
}

type RouteFilter struct {
	sync.Once
	Echo             *echo.Echo
	registeredRoutes map[string]bool
}

func (f *RouteFilter) init() {
	f.registeredRoutes = map[string]bool{}

	for _, route := range f.Echo.Routes() {
		f.registeredRoutes[route.Path] = true
	}
}

func (f *RouteFilter) GetFilteredPath(c echo.Context) string {
	f.Once.Do(f.init)

	path := c.Path()
	if f.registeredRoutes[path] {
		return path
	}

	return "unregistered"
}
