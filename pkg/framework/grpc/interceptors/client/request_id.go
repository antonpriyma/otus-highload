package client

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/framework/grpc/interceptors/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewUnaryClientRequestIDInterceptor(requestIDGetter func(ctx context.Context) string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string, req,
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) (err error) {
		if reqID := requestIDGetter(ctx); reqID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, server.RequestIDHeader, reqID)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
