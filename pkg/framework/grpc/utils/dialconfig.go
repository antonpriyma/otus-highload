package utils

import (
	"context"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/grpc/auth"
	client_interceptors "github.com/antonpriyma/otus-highload/pkg/framework/grpc/interceptors/client"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"

	"google.golang.org/grpc"
)

type DialConfig struct {
	Insecure    bool                           `mapstructure:"insecure"`
	WithBlock   bool                           `mapstructure:"with_block"`
	Credentials RPCCredentialsConfig           `mapstructure:"credentials"`
	Log         client_interceptors.LogParams  `mapstructure:"log"`
	Stat        client_interceptors.StatConfig `mapstructure:"stat"`
}

type RPCCredentialsConfig struct {
	Token        string `mapstructure:"token" json:"-"`
	HeaderName   string `mapstructure:"header_name"`
	WithSecurity bool   `mapstructure:"with_security"` // require transport security
}

func NewDialOptions(
	cfg DialConfig,
	logger log.Logger,
	stat stat.Registry,
	optionalInterceptors ...grpc.UnaryClientInterceptor,
) []grpc.DialOption {
	var opts []grpc.DialOption

	if cfg.WithBlock {
		opts = append(opts, grpc.WithBlock())
	}

	if cfg.Insecure {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithPerRPCCredentials(
		auth.NewPerRPCCredentials(
			cfg.Credentials.Token,
			cfg.Credentials.HeaderName,
			cfg.Credentials.WithSecurity,
		),
	))

	opts = append(opts,
		grpc.WithChainUnaryInterceptor(
			client_interceptors.NewUnaryClientLoggingInterceptor(logger, cfg.Log),
			client_interceptors.NewUnaryClientStatInterceptor(cfg.Stat, stat),
		),
	)

	opts = append(opts, grpc.WithChainUnaryInterceptor(optionalInterceptors...))

	return opts
}

type ConnConfig struct {
	DialConfig  `mapstructure:",squash"`
	URL         string        `mapstructure:"url"`
	InitTimeout time.Duration `mapstructure:"init_timeout"`
}

func NewConn(
	cfg ConnConfig,
	logger log.Logger,
	registry stat.Registry,
	optionalInterceptors ...grpc.UnaryClientInterceptor,
) (*grpc.ClientConn, error) {
	opts := NewDialOptions(cfg.DialConfig, logger, registry, optionalInterceptors...)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cfg.InitTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.URL, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial %s", cfg.URL)
	}

	return conn, nil
}
