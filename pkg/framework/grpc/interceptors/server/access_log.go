package server

import (
	"context"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func NewAccessLogInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		defer func() {
			fields := log.Fields{
				"duration": time.Since(start).String(),
				"method":   info.FullMethod,
			}

			md, ok := metadata.FromIncomingContext(ctx)
			if ok {
				fields["user_agent"] = md.Get("user-agent")
				fields["content-type"] = md.Get("content-type")
				fields["status"] = status.Code(err)
			}

			pr, ok := peer.FromContext(ctx)
			if ok {
				fields["user_ip"] = pr.Addr.String()
			}

			logger.ForCtx(ctx).WithFields(fields).Info("access log")
		}()

		return handler(ctx, req)
	}
}
