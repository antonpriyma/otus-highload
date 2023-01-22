package server

import (
	"context"

	"google.golang.org/grpc"
)

type DisabledHandlers map[string]bool

func DisabledHandlersList(handlers ...string) DisabledHandlers {
	ret := make(map[string]bool, len(handlers))

	for _, handler := range handlers {
		ret[handler] = true
	}

	return ret
}

func NewOptionalInterceptor(
	interceptor grpc.UnaryServerInterceptor,
	disabledHandlers DisabledHandlers,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if disabledHandlers[info.FullMethod] {
			return handler(ctx, req)
		}

		return interceptor(ctx, req, info, handler)
	}
}
