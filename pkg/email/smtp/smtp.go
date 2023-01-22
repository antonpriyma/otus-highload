package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/email"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"

	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset" // Handle most common charsets

	"github.com/rs/xid"
)

const (
	headerMessageID = "Message-ID"
	headerSubject   = "Subject"
	headerReplyTo   = "Reply-To"
	headerFrom      = "From"
	headerTo        = "To"
	headerDate      = "Date"
)

type Auth struct {
	User     string `mapstructure:"user" json:"-"`
	Password string `mapstructure:"password" json:"-"`
}

func (a Auth) isEmpty() bool {
	return a.User == "" && a.Password == ""
}

type Config struct {
	Host               string `mapstructure:"host"`
	Port               int    `mapstructure:"port"`
	SkipTLS            bool   `mapstructure:"skip_tls"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`

	Auth Auth `mapstructure:"auth"`
}

type emailSender struct {
	Auth               smtp.Auth
	Host               string
	Port               int
	SkipTLS            bool
	InsecureSkipVerify bool
	ReplyTo            string
	ServiceHeaders     map[string][]string
	Logger             log.Logger
	Stat               smtpStat

	client *smtp.Client
}

func (e emailSender) HostWithPort() string {
	return fmt.Sprintf("%s:%d", e.Host, e.Port)
}

type smtpStat struct {
	SendDuration stat.TimerCtor `labels:"status"`
}

func NewSender(
	config Config,
	serviceHeaders map[string][]string,
	logger log.Logger,
	registry stat.Registry,
) email.Sender {
	ret := emailSender{
		Host:               config.Host,
		Port:               config.Port,
		SkipTLS:            config.SkipTLS,
		InsecureSkipVerify: config.InsecureSkipVerify,
		Logger:             logger,
		ServiceHeaders:     serviceHeaders,
	}

	if !config.Auth.isEmpty() {
		ret.Auth = smtp.PlainAuth("", config.Auth.User, config.Auth.Password, config.Host)
	}

	stat.NewRegistrar(registry.ForSubsystem("smtp")).MustRegister(&ret.Stat)

	return ret
}

func (e emailSender) Send(
	ctx context.Context,
	to []string,
	subject string,
	bodyPart email.Part,
	opts email.SendOpts,
) (err error) {
	timer := e.Stat.SendDuration.Timer(ctx).Start()
	defer func() {
		timer.WithLabels(stat.Labels{"status": stat.TypedErrorLabel(ctx, err)}).Stop()
	}()

	if opts.EmailFrom == "" {
		return errors.New("empty EmailFrom")
	}

	headers, err := e.prepareHeaders(bodyPart.Header(), to, subject, opts)
	if err != nil {
		return errors.Wrap(err, "failed to prepare headers")
	}

	body, err := e.prepareBody(bodyPart, headers)
	if err != nil {
		return errors.Wrap(err, "failed to prepare body")
	}

	if err = e.validateAndSendEmail(ctx, to, body, opts); err != nil {
		return errors.Wrap(typedSMTPError(err), "failed to validate and send email")
	}

	return nil
}

func (e emailSender) validateAndSendEmail(ctx context.Context, to []string, msg []byte, opts email.SendOpts) error {
	for _, address := range append(to, opts.EmailFrom) {
		if err := validateEmail(address); err != nil {
			err = errors.Wrapf(err, "failed to validate address %q as email", address)
			return errors.Transform(err, ErrBadAddress)
		}
	}

	return e.sendEmail(ctx, to, msg, opts)
}

func (e emailSender) sendEmail(ctx context.Context, to []string, msg []byte, opts email.SendOpts) error {
	client, err := e.newClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to init client")
	}

	e.client = client

	hostname, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "failed to get current hostname")
	}

	if err = e.client.Hello(hostname); err != nil {
		return errors.Wrapf(err, "failed to send hello as %s", hostname)
	}

	if err = e.applyTLS(); err != nil {
		return errors.Wrap(err, "failed to apply tls")
	}

	if err = e.applyAuth(); err != nil {
		return errors.Wrap(err, "failed to apply auth")
	}

	if err = e.client.Mail(opts.EmailFrom); err != nil {
		return errors.Wrapf(err, "failed to send MAIL with %s", opts.EmailFrom)
	}

	for _, recipient := range to {
		if err = e.client.Rcpt(recipient); err != nil {
			return errors.Wrapf(err, "failed to send RCPT with %s", recipient)
		}
	}

	w, err := e.client.Data()
	if err != nil {
		return errors.Wrapf(err, "failed to send DATA to %s", e.Host)
	}

	if _, err = w.Write(msg); err != nil {
		return errors.Wrapf(err, "failed to write message to %s", e.Host)
	}

	if err = w.Close(); err != nil {
		return errors.Wrap(err, "failed to close write")
	}

	return errors.Wrapf(e.client.Quit(), "failed to send QUIT to %s", e.Host)
}

func (e emailSender) newClient(ctx context.Context) (*smtp.Client, error) {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", e.HostWithPort())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial %s", e.HostWithPort())
	}

	client, err := smtp.NewClient(conn, e.Host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init smtp client")
	}

	return client, nil
}

func (e emailSender) applyTLS() error {
	if e.SkipTLS {
		return nil
	}

	if serverSupportTLS, _ := e.client.Extension("STARTTLS"); !serverSupportTLS {
		return email.ErrTLSUnsupported
	}

	config := &tls.Config{
		ServerName: e.Host,
		MinVersion: tls.VersionTLS12,
		//nolint: gosec
		InsecureSkipVerify: e.InsecureSkipVerify,
	}

	if err := e.client.StartTLS(config); err != nil {
		return errors.Transform(err, email.ErrFailedStartTLS)
	}

	return nil
}

func (e emailSender) applyAuth() error {
	if e.Auth == nil {
		return nil
	}

	if serverSupportAuth, _ := e.client.Extension("AUTH"); !serverSupportAuth {
		return nil
	}

	if err := e.client.Auth(e.Auth); err != nil {
		return errors.Transform(err, email.ErrAuthFailed)
	}

	return nil
}

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return errors.Wrap(err, "failed to parse email")
}

func (e emailSender) prepareHeaders(
	bodyHeaders email.Header,
	to []string,
	subject string,
	opts email.SendOpts,
) (message.Header, error) {
	var headers message.Header

	bodyHeaders.ExtendMessageHeaders(&headers)
	email.Header(e.ServiceHeaders).ExtendMessageHeaders(&headers)

	headers.Set(headerMessageID, messageIDHeader(opts.EmailFrom))
	headers.Set(headerReplyTo, addressHeader(opts.ReplyTo, ""))
	headers.Set(headerFrom, addressHeader(opts.EmailFrom, opts.NameFrom))
	headers.Set(headerSubject, mime.BEncoding.Encode("UTF-8", subject))
	headers.Set(headerDate, time.Now().Format(time.RFC1123Z))

	for _, receiver := range to {
		headers.Add(headerTo, addressHeader(receiver, ""))
	}

	err := email.ValidateHeaders(&headers)
	if err != nil {
		return message.Header{}, errors.Wrap(err, "failed to validate body headers")
	}

	return headers, nil
}

func (e emailSender) prepareBody(
	body email.Part,
	headers message.Header,
) ([]byte, error) {
	var b bytes.Buffer

	msg, err := message.CreateWriter(&b, headers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create writer")
	}
	if err = body.MarshalEML(msg); err != nil {
		return nil, errors.Wrap(err, "failed to marshal EML")
	}
	if err = msg.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close msg writer")
	}

	return b.Bytes(), nil
}

func addressHeader(email, name string) string {
	if name == "" {
		return email
	}
	return (&mail.Address{Name: name, Address: email}).String()
}

var messageIDReplacer = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func messageIDHeader(sender string) string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "local"
	}

	sender = messageIDReplacer.ReplaceAllString(sender, "-")
	sender = strings.Trim(sender, "-")

	return fmt.Sprintf("<%s.%s@%s>", xid.New().String(), sender, hostname)
}
