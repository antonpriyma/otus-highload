package middleware

import (
	"github.com/antonpriyma/otus-highload/internal/pkg/contextlib"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/labstack/echo"
)

const XUserID = "X-User-ID"

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Request().Header.Get(XUserID)

		if _, err := echoutils.GetContext(c); err != nil {
			echoutils.StoreContext(c.Request().Context(), c)
		}
		echoutils.StoreContext(contextlib.WithUserID(echoutils.MustGetContext(c), userID), c)
		return next(c)
	}
}
