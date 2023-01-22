package server

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/debug"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorHandler func(ctx context.Context, info *grpc.UnaryServerInfo, err error) error

func WrapInterceptorsWithErrorHandler(
	errorHandler ErrorHandler,
	interceptors ...grpc.UnaryServerInterceptor,
) []grpc.UnaryServerInterceptor {
	helper := errorHelper{}

	wrapped := make([]grpc.UnaryServerInterceptor, 0, len(interceptors)+1)
	wrapped = append(wrapped, makeInterceptorForErrorHandler(errorHandler, helper))

	for _, interceptor := range interceptors {
		wrapped = append(wrapped, helper.wrapInterceptor(interceptor))
	}

	return wrapped
}

func makeInterceptorForErrorHandler(
	errorHandler ErrorHandler,
	helper errorHelper,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx = helper.syncContext(ctx)

		resp, err := handler(ctx, req)
		if err != nil {
			return resp, errorHandler(helper.lastContext(ctx), info, err)
		}

		return resp, nil
	}
}

type errorHelper struct{}

func (h errorHelper) wrapInterceptor(interceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		return interceptor(ctx, req, info, h.wrapHandler(handler))
	}
}

func (h errorHelper) wrapHandler(handler grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		ctx = h.syncContext(ctx)
		return handler(ctx, req)
	}
}

type errorHandlerStorageKey struct{}

type errorHandlerStorage struct {
	Context context.Context
}

func (h errorHelper) syncContext(ctx context.Context) context.Context {
	stored, ok := ctx.Value(errorHandlerStorageKey{}).(*errorHandlerStorage)
	if !ok {
		stored = &errorHandlerStorage{}
		ctx = context.WithValue(ctx, errorHandlerStorageKey{}, stored)
	}

	stored.Context = ctx
	return ctx
}

func (h errorHelper) lastContext(ctx context.Context) context.Context {
	stored, ok := ctx.Value(errorHandlerStorageKey{}).(*errorHandlerStorage)
	if !ok {
		return ctx
	}

	return stored.Context
}

type CallbackFunc func(ctx context.Context, status *status.Status)

type DefaultErrorHandler struct {
	Logger   log.Logger
	Callback CallbackFunc
}

func (e DefaultErrorHandler) HandleError(
	ctx context.Context,
	info *grpc.UnaryServerInfo,
	err error,
) error {
	logger := e.Logger.ForCtx(ctx).WithError(err)
	if errors.Is(err, errors.ErrPanic) {
		logger = logger.WithField("stack", debug.ErrorStackTrace(err))
	}

	logger.WithFields(log.Fields{
		"response": err.Error(),
		"method":   info.FullMethod,
	}).Error("error happened during request")

	if status.Code(err) == codes.Unknown {
		err = status.Error(codes.Internal, err.Error())
	}

	if e.Callback != nil {
		e.Callback(ctx, status.Convert(err))
	}

	return err
}
