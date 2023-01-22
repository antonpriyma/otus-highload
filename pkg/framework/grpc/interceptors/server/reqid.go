package server

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/context/reqid"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const RequestIDHeader = "X-Request-ID"

func NewRequestIDInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		reqID := getReqIDFromMetadata(ctx)
		if reqID == "" {
			reqID = reqid.GenerateRequestID()
		}

		ctx = reqid.SetRequestID(ctx, reqID)

		header := metadata.Pairs(RequestIDHeader, reqID)
		if err := grpc.SetHeader(ctx, header); err != nil {
			logger.ForCtx(ctx).WithError(err).Error("unable to set %s header", RequestIDHeader)
		}

		ctx = log.AddCtxFields(ctx, log.Fields{
			"request_id": reqID,
		})

		return handler(ctx, req)
	}
}

func getReqIDFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	headers := md.Get(RequestIDHeader)
	if len(headers) == 0 {
		return ""
	}

	return headers[0]
}
