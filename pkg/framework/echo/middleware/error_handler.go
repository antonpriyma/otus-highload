package middleware

import (
	"context"
	"net/http"

	"github.com/antonpriyma/otus-highload/pkg/debug"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoerrors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/antonpriyma/otus-highload/pkg/stat/stub"

	"github.com/labstack/echo"
)

type CallbackFunc func(ectx echo.Context, err *echo.HTTPError)

type ErrorHandler struct {
	Echo         *echo.Echo
	Logger       log.Logger
	Callback     CallbackFunc
	StatRegistry stat.Registry
}

func (e ErrorHandler) NewHandlerFunc() func(err error, ectx echo.Context) {
	statReg := stub.NewStubRegistry()
	if e.StatRegistry != nil {
		statReg = e.StatRegistry
	}

	var errorHandlerStat struct {
		ResponseErrors stat.CounterCtor `labels:"api_status,cause_type"`
	}
	stat.NewRegistrar(statReg.ForSubsystem("error_handler")).MustRegister(&errorHandlerStat)

	return func(err error, ectx echo.Context) {
		ctx := ectx.Request().Context()

		defer func() {
			if r := errors.RecoverError(recover()); r != nil {
				e.Logger.
					ForCtx(ctx).
					WithError(r).
					WithField("request_path", ectx.Path()).
					WithField("stack", debug.ErrorStackTrace(r)).
					Error("panic while handling error")
			}
		}()

		httpErr := extractHTTPError(ctx, err)

		logger := e.Logger.ForCtx(ctx).WithError(err).WithField("response", httpErr.Error())
		if errors.Is(err, errors.ErrPanic) {
			logger = logger.WithField("stack", debug.ErrorStackTrace(err))
		}

		switch httpErr.Code {
		case http.StatusInternalServerError:
			logger.Error("error happened during request")
		case http.StatusUnauthorized, http.StatusForbidden:
			logger.Info("error happened during request")
		default:
			logger.Warn("error happened during request")
		}

		if e.Callback != nil {
			e.Callback(ectx, httpErr)
		}
		e.Echo.DefaultHTTPErrorHandler(httpErr, ectx)

		// order is important, counter should be after DefaultHTTPErrorHandler
		// because before we don't have response status
		errorHandlerStat.ResponseErrors.Counter(ctx).WithLabels(stat.Labels{
			"api_status": echoutils.GetResponseStatus(ectx),
			"cause_type": stat.TypedErrorLabel(ctx, err),
		}).Add(1)
	}
}

func extractHTTPError(
	ctx context.Context,
	err error,
) *echo.HTTPError {
	if ctx.Err() != nil {
		return echoerrors.CanceledError(err).HTTPError()
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	var respErr echoerrors.ResponseError
	if errors.As(err, &respErr) {
		return respErr.HTTPError()
	}

	return echoerrors.InternalError(err).HTTPError()
}
