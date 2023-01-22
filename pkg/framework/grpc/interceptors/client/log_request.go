package client

import (
	"context"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type LogParams struct {
	Debug   bool     `mapstructure:"debug"`
	Headers []string `mapstructure:"headers"`
}

func NewUnaryClientLoggingInterceptor(logger log.Logger, params LogParams) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req,
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {

		startTime := time.Now()

		var headers metadata.MD

		opts = append(opts, grpc.Header(&headers))

		err := invoker(ctx, method, req, reply, cc, opts...)

		fields := log.Fields{
			"client_request_method":    method,
			"client_request_resp_code": status.Code(err),
			"client_request_duration":  time.Since(startTime),
			"client_request_addr":      cc.Target(),

			"request_headers":  filterHeaders(getHeadersFromContext(ctx), params.Headers),
			"response_headers": headers,
		}

		if params.Debug {
			fields.Extend(log.Fields{
				"request_body":  req,
				"response_body": reply,
			})
		}

		if err != nil {
			fields.Extend(log.Fields{
				"response_error": err,
				"response_code":  status.Code(err).String(),
			})
		}

		logger.ForCtx(ctx).WithFields(fields).Info("request")
		return err
	}
}

func getHeadersFromContext(ctx context.Context) metadata.MD {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return metadata.MD{}
	}

	return md
}

func filterHeaders(md metadata.MD, headers []string) metadata.MD {
	ret := make(metadata.MD, len(headers))
	for _, header := range headers {
		val := md.Get(header)
		if len(val) > 0 {
			ret[header] = val
		}
	}

	return ret
}
