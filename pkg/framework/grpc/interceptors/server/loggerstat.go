package server

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat/loggerstat"
	"google.golang.org/grpc"
)

func NewLoggerStatInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		ctx = loggerstat.InitStatForCtx(ctx)
		defer loggerstat.PrintStat(ctx, logger)

		return handler(ctx, req)
	}
}
