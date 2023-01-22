package auth

import (
	"context"

	"google.golang.org/grpc/credentials"
)

type tokenAuth struct {
	Token      string
	HeaderName string
	Security   bool
}

func NewPerRPCCredentials(token, headerName string, withSecurity bool) credentials.PerRPCCredentials {
	return tokenAuth{
		Token:      token,
		HeaderName: headerName,
		Security:   withSecurity,
	}
}

func (t tokenAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		t.HeaderName: t.Token,
	}, nil
}

func (t tokenAuth) RequireTransportSecurity() bool {
	return t.Security
}
