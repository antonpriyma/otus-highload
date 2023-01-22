package server

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"google.golang.org/grpc"
)

func RecoverInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		r := errors.RecoverError(recover())
		if r != nil {
			err = r
		}
	}()

	return handler(ctx, req)
}
