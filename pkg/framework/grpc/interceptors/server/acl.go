package server

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrUnauthorized     = status.Error(codes.Unauthenticated, "acl unauthorized")
	errACLMatchNotFound = utils.NewTypedError("acl_match_not_found", "acl match not found")
	errMethodNowAllowed = utils.NewTypedError("acl_method_not_allowed", "method not allowed")
	errEmptyToken       = utils.NewTypedError("acl_empty_serverside_token", "empty serverside token")
)

type ACLConfig struct {
	HeaderName string    `mapstructure:"header_name"`
	Enabled    bool      `mapstructure:"enabled"`
	Nodes      []ACLNode `mapstructure:"nodes"`
}

type ACLNode struct {
	Owner   string   `mapstructure:"owner"`
	Token   string   `mapstructure:"token" json:"-"`
	Methods []string `mapstructure:"methods"`
}

type serversideACL struct {
	Config ACLConfig
	Logger log.Logger
}

func NewServersideACLInterceptor(cfg ACLConfig, logger log.Logger) grpc.UnaryServerInterceptor {
	return serversideACL{
		Config: cfg,
		Logger: logger,
	}.interceptor
}

func (a serversideACL) interceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if !a.Config.Enabled {
		return handler(ctx, req)
	}

	node, err := a.authorizeRequest(ctx, info)
	if err != nil {
		return nil, errors.Transform(err, ErrUnauthorized)
	}

	ctx = log.AddCtxFields(ctx, map[string]interface{}{"acl_token_owner": node.Owner})

	return handler(ctx, req)
}

func (a serversideACL) authorizeRequest(ctx context.Context, info *grpc.UnaryServerInfo) (ACLNode, error) {
	token := a.getTokenFromMetadata(ctx)
	if token == "" {
		return ACLNode{}, errEmptyToken
	}

	for _, node := range a.Config.Nodes {
		if node.Token == token {
			if utils.ContainsString(node.Methods, info.FullMethod) {
				return node, nil
			}

			return ACLNode{}, errMethodNowAllowed
		}
	}

	return ACLNode{}, errACLMatchNotFound
}

func (a serversideACL) getTokenFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	header := md.Get(a.Config.HeaderName)
	if len(header) == 0 {
		return ""
	}

	return header[0]
}
