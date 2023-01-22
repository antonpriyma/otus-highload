package email

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/utils"
	"github.com/emersion/go-message"
)

type SendOpts struct {
	EmailFrom string
	NameFrom  string
	ReplyTo   string
}

var (
	ErrTLSUnsupported = utils.NewTypedError("tls_unsupported", "tls unsupported by a server")
	ErrFailedStartTLS = utils.NewTypedError("start_tls_failed", "start tls failed")
	ErrAuthFailed     = utils.NewTypedError("auth_failed", "auth failed")
)

type Part interface {
	MarshalEML(writer *message.Writer) error
	Header() Header
}

type Sender interface {
	Send(ctx context.Context, to []string, subject string, body Part, opts SendOpts) error
}
