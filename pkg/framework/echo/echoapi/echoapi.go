package echoapi

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/middleware"
	"github.com/antonpriyma/otus-highload/pkg/framework/service"
	"github.com/antonpriyma/otus-highload/pkg/stat/cmnlabelsstat"
	"github.com/labstack/echo"
)

type Configured interface {
	service.Configured
	APIConfig() Config
}

type Config struct {
	ServeConfig service.ServeConfig `mapstructure:"serve_config"`
	Listen      string              `mapstructure:"listen"`
}

func (c Config) APIConfig() Config {
	return c
}

type ToolSet struct {
	cfg  Config
	echo *echo.Echo

	service.ToolSet
	API *echo.Group
}

func (t ToolSet) Run() {
	service.Serve(
		context.Background(),
		t.Logger,
		t.cfg.ServeConfig,
		&EchoServer{
			echo: t.echo,
			addr: t.cfg.Listen,
		},
	)
}

func (t ToolSet) ErrorHandler() echo.HTTPErrorHandler {
	return t.echo.HTTPErrorHandler
}

func (t ToolSet) SetErrorHandler(h echo.HTTPErrorHandler) {
	t.echo.HTTPErrorHandler = h
}

func (t ToolSet) SetErrorHandlerCallback(callbackFunc middleware.CallbackFunc) {
	t.echo.HTTPErrorHandler = middleware.ErrorHandler{
		Echo:     t.echo,
		Logger:   t.Logger,
		Callback: callbackFunc,
	}.NewHandlerFunc()
}

func New(appCfg Configured) ToolSet {
	toolSet := RawToolSet(appCfg)
	toolSet.Init()

	return toolSet
}

func RawToolSet(appCfg Configured) ToolSet {
	svc := service.New(appCfg)

	return ToolSet{
		cfg: appCfg.APIConfig(),

		ToolSet: svc,
	}
}

func (t *ToolSet) Init() {
	svc := t.ToolSet
	svc.StatRegistry = cmnlabelsstat.NewRegistry(svc.StatRegistry, []string{"api_end", "api_method"})

	e := echo.New()
	e.Use(middleware.Recover(svc.Logger))
	e.HTTPErrorHandler = middleware.ErrorHandler{
		Echo:         e,
		Logger:       svc.Logger,
		StatRegistry: svc.StatRegistry,
	}.NewHandlerFunc()
	e.GET("/metrics", echo.WrapHandler(svc.PromHandler))
	e.GET("/debug/*", echo.WrapHandler(svc.PProfHandler))

	routeFilter := echoutils.NewRouteFilter(e)
	api := e.Group(
		"",
		middleware.SetCtxReqIDMiddleware,
		middleware.LoggerStatMiddleware(svc.Logger),
		middleware.AroundResponseMiddleware(
			middleware.AccessLogAfterware(svc.Logger, routeFilter),
			middleware.StatAroundResponse(svc.StatRegistry, routeFilter),
		),
	)

	t.echo = e
	t.API = api
}
