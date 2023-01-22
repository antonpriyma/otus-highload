package echoutils

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/labstack/echo"
)

const contextKey = "echoutils-context-key"

func MustGetContext(c echo.Context) context.Context {
	ctx, err := GetContext(c)
	if err != nil {
		panic(err)
	}

	return ctx
}

func GetContext(c echo.Context) (context.Context, error) {
	rawCtx := c.Get(contextKey)
	if rawCtx == nil {
		return nil, errors.New("no context in echo context")
	}

	ctx, ok := rawCtx.(context.Context)
	if !ok {
		return nil, errors.Errorf("bad type of stored context, got type %T", rawCtx)
	}

	return ctx, nil
}

func StoreContext(ctx context.Context, c echo.Context) {
	c.Set(contextKey, ctx)
}
