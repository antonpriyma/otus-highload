package reqid

import (
	"context"
	"fmt"
	"math/rand"
)

type requestIDKey struct{}

func GetRequestID(ctx context.Context) string {
	reqID := ctx.Value(requestIDKey{})
	if reqID == nil {
		return ""
	}

	return reqID.(string)
}

func GetOrGenerateRequestID(ctx context.Context) string {
	reqID := ctx.Value(requestIDKey{})
	if reqID == nil {
		return GenerateRequestID()
	}

	return reqID.(string)
}

func SetRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, reqID)
}

func GenerateRequestID() string {
	return fmt.Sprintf("%016x", rand.Int()) // nolint: gosec
}
