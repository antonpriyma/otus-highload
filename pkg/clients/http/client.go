package http

import (
	"context"
	"net/http"
)

//go:generate mockgen -destination=./internal/mock/mock_generated.go -package=mock github.com/antonpriyma/otus-highload/pkg/clients/http Request,RequestWithHeaders,RequestWithBody,Response
//go:generate mockgen -destination=./mock/mock_generated.go -package=mock github.com/antonpriyma/otus-highload/pkg/clients/http Client

type Client interface {
	PerformRequest(ctx context.Context, req Request, res Response) error
}

type Request interface {
	URL() string
	Method() string
}

type RequestWithHeaders interface {
	Request
	Headers() http.Header
}

type RequestWithBody interface {
	Request
	Body() ([]byte, error)
}

type RequestWithRequestID interface {
	Request
	RequestID() string
}

type Response interface {
	ReadFrom(*http.Response) error
}
