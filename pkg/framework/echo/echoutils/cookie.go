package echoutils

import "github.com/labstack/echo"

func HasCookie(c echo.Context, name string) bool {
	_, err := c.Request().Cookie(name)
	return err == nil
}
