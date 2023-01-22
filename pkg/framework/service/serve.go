package service

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/log"
)

type Server interface {
	Init(ctx context.Context) error
	Run(ctx context.Context) error
	Graceful(ctx context.Context) error
	Stop(ctx context.Context) error
}

type ServeConfig struct {
	GracefulWait time.Duration `mapstructure:"graceful_wait"`
	StopWait     time.Duration `mapstructure:"stop_wait"`
}

func Serve(
	ctx context.Context,
	logger log.Logger,
	cfg ServeConfig,
	server Server,
) {
	logger = logger.ForCtx(ctx)

	err := server.Init(ctx)
	if err != nil {
		logger.WithError(err).Fatal("error during initing server")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	runCtx, runCancel := context.WithCancel(ctx)
	go func() {
		runErr := server.Run(runCtx)
		logger.WithError(runErr).Warn("server run stopped")
		close(quit)
	}()

	logger.Warn("server run started")
	<-quit
	logger.Warn("server shutdown inited")

	gracefulCtx, gracefulCancel := context.WithTimeout(ctx, cfg.GracefulWait)
	defer gracefulCancel()

	err = server.Graceful(gracefulCtx)
	if err != nil {
		logger.WithError(err).Error("error during graceful server stopping")
	}

	runCancel()

	stopCtx, stopCancel := context.WithTimeout(ctx, cfg.StopWait)
	defer stopCancel()

	err = server.Stop(stopCtx)
	if err != nil {
		logger.WithError(err).Fatal("error during shutting down server")
	}

	logger.Warn("server shutdowned normally")
}
