package service

import (
	"context"
	"net"
	"net/http"

	"github.com/antonpriyma/otus-highload/pkg/framework/echo/middleware"
	"github.com/antonpriyma/otus-highload/pkg/framework/service"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Configured interface {
	service.Configured
	ServerConfig() Config
}

var _ service.Server = &ToolSet{}

type Config struct {
	Listen           string `mapstructure:"listen"`
	PrometheusListen string `mapstructure:"prometheus_listen"`
}

func New(appCfg Configured) *ToolSet {
	// should be inited before others cause of config parsing
	baseTool := service.New(appCfg)

	return &ToolSet{
		config:  appCfg.ServerConfig(),
		ToolSet: baseTool,
		stopCh:  make(chan struct{}),
	}
}

type ToolSet struct {
	config   Config
	listener net.Listener
	server   *grpc.Server
	echo     *echo.Echo // using for probes and probes
	stopCh   chan struct{}

	service.ToolSet
}

func (t *ToolSet) Graceful(ctx context.Context) error {
	go func() {
		t.server.GracefulStop()
		t.stopCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return errors.New("failed to graceful shutdown")
	case <-t.stopCh:
		return nil
	}
}

func (t *ToolSet) SetServer(s *grpc.Server) {
	t.server = s
}

func (t *ToolSet) Init(ctx context.Context) error {
	ln, err := net.Listen("tcp", t.config.Listen)
	if err != nil {
		return errors.Wrap(err, "failed to bind address")
	}
	t.listener = ln

	e := echo.New()
	e.Use(middleware.Recover(t.Logger))

	e.HTTPErrorHandler = middleware.ErrorHandler{
		Echo:   e,
		Logger: t.Logger,
	}.NewHandlerFunc()
	e.GET("/metrics", echo.WrapHandler(t.PromHandler))
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	e.GET("/debug/*", echo.WrapHandler(t.PProfHandler))

	echoLn, err := net.Listen("tcp", t.config.PrometheusListen)
	if err != nil {
		return errors.Wrap(err, "failed to bind prometheus port")
	}
	e.Listener = echoLn

	t.echo = e

	return nil
}

func (t *ToolSet) Run(ctx context.Context) error {
	// launch prometheus
	go func() {
		err := t.echo.Start(t.config.PrometheusListen)
		t.Logger.WithError(err).Warn("shutting down prometheus server")
	}()

	return t.server.Serve(t.listener)
}

func (t *ToolSet) Stop(ctx context.Context) error {
	go func() {
		t.server.Stop()
		t.stopCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return errors.New("failed to shutdown")
	case <-t.stopCh:
		return nil
	}
}
