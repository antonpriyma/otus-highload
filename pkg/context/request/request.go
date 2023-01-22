package request

import (
	"context"
	"net/http"
)

type requestKey struct{}

func GetRequest(ctx context.Context) *http.Request {
	req := ctx.Value(requestKey{})
	if req == nil {
		return nil
	}

	return req.(*http.Request)
}

func MustGetRequest(ctx context.Context) *http.Request {
	req := GetRequest(ctx)
	if req == nil {
		panic("there is no http.Request in context")
	}

	return req
}

func SetRequest(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, requestKey{}, req)
}
