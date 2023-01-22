package stat

import (
	"context"
	"net"

	"github.com/antonpriyma/otus-highload/pkg/errors"
)

func TypedErrorLabel(ctx context.Context, err error) string {
	source := getSourceFromError(err)
	label := getLabelFromError(ctx, err)
	if source == "" {
		return label
	}
	return source + "." + label
}

func getLabelFromError(ctx context.Context, err error) string {
	if err == nil {
		return "ok"
	}
	if ctx.Err() != nil {
		return "context_canceled"
	}

	var typedErr errors.TypedError
	if ok := errors.As(err, &typedErr); ok {
		return typedErr.ErrorType()
	}

	var netErr net.Error
	if ok := errors.As(err, &netErr); ok {
		return netErrorLabel(netErr)
	}

	return "fail"
}

func getSourceFromError(err error) string {
	var errWithSource *errors.ErrorWithSource
	if ok := errors.As(err, &errWithSource); ok {
		return errWithSource.GetSource()
	}
	return ""
}

func netErrorLabel(err net.Error) string {
	label := "net"

	var opErr *net.OpError
	if ok := errors.As(err, &opErr); ok {
		label += "_" + opErr.Op
	}

	switch {
	case err.Timeout():
		label += "_timeout"
	default:
		label += "_fail"
	}

	return label
}
