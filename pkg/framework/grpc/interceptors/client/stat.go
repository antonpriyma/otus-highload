package client

import (
	"context"
	"fmt"

	"github.com/antonpriyma/otus-highload/pkg/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type StatConfig struct {
	Service string `mapstructure:"service"`
}

func NewUnaryClientStatInterceptor(cfg StatConfig, registry stat.Registry) grpc.UnaryClientInterceptor {
	var statSender struct {
		Duration stat.TimerCtor `labels:"method,status"`
	}

	stat.NewRegistrar(registry.ForSubsystem(fmt.Sprintf("grpc_client_%s", cfg.Service))).MustRegister(&statSender)

	return func(
		ctx context.Context,
		method string, req,
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) (err error) {
		timer := statSender.Duration.Timer(ctx).Start()

		defer func() {
			timer.WithLabels(stat.Labels{
				"method": method,
				"status": status.Code(err).String(),
			}).Stop()
		}()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
