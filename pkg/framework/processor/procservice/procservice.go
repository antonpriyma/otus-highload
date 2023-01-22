package procservice

import (
	"context"
	"net"
	"net/http"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/middleware"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor/processorpool"
	"github.com/antonpriyma/otus-highload/pkg/framework/service"

	"github.com/labstack/echo"
)

type Configured interface {
	service.Configured
	ProcessorConfig() Config
}

type Config struct {
	PrometheusListen string                   `mapstructure:"prometheus_listen"`
	Pool             processorpool.PoolConfig `mapstructure:"pool"`
}

func (c Config) ProcessorConfig() Config {
	return c
}

func New(appCfg Configured) *ToolSet {
	// should be inited before others cause of config parsing
	baseTool := service.New(appCfg)

	return &ToolSet{
		config:  appCfg.ProcessorConfig(),
		ToolSet: baseTool,
	}
}

var _ service.Server = &ToolSet{}

type ToolSet struct {
	config Config
	echo   *echo.Echo
	pool   processorpool.Pool

	service.ToolSet
}

func (t *ToolSet) SetProcessor(
	taskGetter processor.TaskGetter,
	middlewares []processor.MiddlewareFunc,
	errorHandler processor.ErrorHandler,
) {
	app := processorpool.AppConfig{
		Task:         taskGetter,
		Middlewares:  middlewares,
		ErrorHandler: errorHandler,
	}

	t.pool = processorpool.New(app, t.config.Pool, t.Logger, t.StatRegistry)
}

func (t *ToolSet) Init(ctx context.Context) error {
	if t.pool == nil {
		return errors.New("processor was not set")
	}

	svc := t.ToolSet

	e := echo.New()
	e.Use(middleware.Recover(svc.Logger))
	e.HTTPErrorHandler = middleware.ErrorHandler{
		Echo:   e,
		Logger: svc.Logger,
	}.NewHandlerFunc()
	e.GET("/metrics", echo.WrapHandler(svc.PromHandler))
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	e.GET("/debug/*", echo.WrapHandler(svc.PProfHandler))

	ln, err := net.Listen("tcp", t.config.PrometheusListen)
	if err != nil {
		return errors.Wrap(err, "failed to bind prometheus port")
	}
	e.Listener = ln

	t.echo = e

	return nil
}

func (t ToolSet) Run(ctx context.Context) error {
	// launch prometheus
	go func() {
		err := t.echo.Start(t.config.PrometheusListen)
		t.Logger.WithError(err).Warn("shutting down prometheus server")
	}()

	// run processor
	t.pool.Run(ctx)

	return nil
}

func (t ToolSet) Graceful(ctx context.Context) error {
	return t.pool.Graceful(ctx)
}

func (t ToolSet) Stop(ctx context.Context) error {
	return t.pool.Stop(ctx)
}
