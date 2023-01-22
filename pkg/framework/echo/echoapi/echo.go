package echoapi

import (
	"context"
	"net"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/labstack/echo"
)

type EchoServer struct {
	echo *echo.Echo
	addr string

	healthChecker *healthChecker
}

func NewEchoServer(addr string) *EchoServer {
	return &EchoServer{
		echo: echo.New(),
		addr: addr,
	}
}

func (e *EchoServer) Init(_ context.Context) error {
	e.echo.HideBanner = true
	e.echo.HidePort = true

	e.healthChecker = newHealthChecker(true)
	e.echo.GET("/ping", e.healthChecker.HTTPHandler)

	ln, err := net.Listen("tcp", e.addr)
	if err != nil {
		return errors.Wrap(err, "failed to bind addr")
	}

	e.echo.Listener = ln

	return nil
}

func (e EchoServer) Run(_ context.Context) error {
	return e.echo.Start(e.addr)
}

func (e EchoServer) Graceful(ctx context.Context) error {
	e.healthChecker.ChangeHealthStatus(false)

	<-ctx.Done()
	return nil
}

func (e EchoServer) Stop(ctx context.Context) error {
	return e.echo.Shutdown(ctx)
}
