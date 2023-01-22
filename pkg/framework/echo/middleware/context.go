package middleware

import (
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/labstack/echo"
)

func ContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		echoutils.StoreContext(c.Request().Context(), c)

		return next(c)
	}
}
