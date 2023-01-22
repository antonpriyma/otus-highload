package server

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func NewStatInterceptor(registry stat.Registry) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var statSender struct {
			RequestDuration stat.TimerCtor `labels:"status,method"`
		}
		stat.NewRegistrar(registry.ForSubsystem("middleware")).MustRegister(&statSender)

		timer := statSender.RequestDuration.Timer(ctx).WithLabels(stat.Labels{
			"method": info.FullMethod,
		}).Start()

		resp, err := handler(ctx, req)

		timer.WithLabels(stat.Labels{"status": status.Code(err).String()}).Stop()

		return resp, err
	}
}
