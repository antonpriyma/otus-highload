package middleware

import (
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/labstack/echo"
)

func Recover(logger log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				rec := errors.RecoverError(recover())
				if rec != nil {
					err = rec
				}
			}()

			return next(c)
		}
	}
}
