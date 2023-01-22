package middleware

import "github.com/labstack/echo"

type AfterFunc func(echo.Context)

type BeforeFunc func(echo.Context) AfterFunc

func AroundResponseMiddleware(fns ...BeforeFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, fn := range fns {
				defer fn(c)(c)
			}

			if err := next(c); err != nil {
				c.Error(err)
			}

			return nil
		}
	}
}
