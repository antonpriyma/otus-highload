package service

import (
	"context"
	"flag"
	"net/http"
	"net/http/pprof"

	"github.com/antonpriyma/otus-highload/pkg/framework/config"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/antonpriyma/otus-highload/pkg/stat/combinedstat"
	"github.com/antonpriyma/otus-highload/pkg/stat/loggerstat"
	"github.com/antonpriyma/otus-highload/pkg/stat/prometheus"
	"github.com/antonpriyma/otus-highload/pkg/utils"
)

type Configured interface {
	ServiceConfig() *Config
}

type Config struct {
	Log     log.Config `mapstructure:"log"`
	Version Version
}

func (c *Config) ServiceConfig() *Config {
	return c
}

type ToolSet struct {
	cfg Config

	Context      context.Context
	Logger       log.Logger
	StatRegistry stat.Registry
	PromHandler  http.Handler
	PProfHandler http.Handler
}

func New(appCfg Configured) ToolSet {
	flag.Parse()

	ctx := context.Background()
	ctx = loggerstat.InitDummyForCtx(ctx)
	ctx = log.AddCtxFields(ctx, log.Fields{"thread": "main"})

	logger := log.Default().ForCtx(ctx)

	viperCfg, err := config.NewConfig()
	utils.Must(logger, err, "failed to init config")

	err = viperCfg.Unmarshal(appCfg)
	utils.Must(logger, err, "failed to unmarshal config")

	cfg := appCfg.ServiceConfig()
	cfg.Version = Version{
		Dist:    Dist,
		Release: Release,
	}

	logger.WithField("config", appCfg).Warn("have read config")

	configuredLogger, err := log.NewLogrusLogger(cfg.Log)
	utils.Must(logger, err, "failed to init logger")
	logger = configuredLogger.ForCtx(ctx)

	promRegistry, promHTTP := prometheus.NewRegistry(cfg.Log.App)
	loggerStatRegistry := loggerstat.NewRegistry(cfg.Log.App, logger)
	statRegistry := combinedstat.NewRegistry(promRegistry, loggerStatRegistry)
	pprofHandler := createPProfHandler()

	return ToolSet{
		cfg:          *cfg,
		Context:      ctx,
		Logger:       logger,
		StatRegistry: statRegistry,
		PromHandler:  promHTTP,
		PProfHandler: pprofHandler,
	}
}

func createPProfHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}
