package echoutils

import (
	"github.com/antonpriyma/otus-highload/pkg/clients/sentry"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoerrors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/labstack/echo"
)

type Validatable interface {
	Validate() error
}

func validate(obj interface{}) error {
	validatableObj, ok := obj.(Validatable)
	if !ok {
		return nil
	}

	return validatableObj.Validate()
}

func Bind(c echo.Context, logger log.Logger, val interface{}) error {
	ctx := c.Request().Context()

	err := c.Bind(val)
	if err != nil {
		return echoerrors.ValidationError(
			errors.Wrapf(err, "failed to bind data: %v", err),
			"bad request",
			echoerrors.ValidationErrorFields{},
		)
	}

	sentry.AddContextExtra(ctx, "incoming request", val)
	c.SetRequest(c.Request().WithContext(ctx))

	logger.ForCtx(ctx).Debugf("incoming request value: %+v", val)

	err = validate(val)
	if err != nil {
		return errors.Wrap(err, "validation error")
	}

	return nil
}
