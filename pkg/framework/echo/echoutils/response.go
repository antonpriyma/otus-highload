package echoutils

import (
	"fmt"

	"github.com/labstack/echo"
)

func GetResponseStatus(c echo.Context) string {
	if c.Response() != nil {
		return fmt.Sprintf("%d", c.Response().Status)
	}

	return "unknown"
}
