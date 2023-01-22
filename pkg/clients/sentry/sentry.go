package sentry

import (
	"context"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/getsentry/sentry-go"
)

type Config struct {
	DSN string `mapstructure:"dsn" json:"-"`

	CutHeaders []string `mapstructure:"cut_headers"`

	AppName     string `mapstructure:"app_name"`
	Environment string `mapstructure:"environment"`
	Release     string `mapstructure:"-"`
	Dist        string `mapstructure:"-"`

	Debug   bool `mapstructure:"debug"`
	Enabled bool `mapstructure:"enabled"`

	TransportBufferSize int           `mapstructure:"transport_buffer_size"`
	TransportTimeout    time.Duration `mapstructure:"transport_timeout"`

	SampleRate map[Level]float64 `mapstructure:"sample_rate"`
}

type Client struct {
	Config     Config
	CutHeaders map[string]bool
	Hub        *sentry.Hub
	Rand       *rand.Rand

	defaultTags map[string]string
}

func NewClient(cfg Config) (Client, error) {
	if !cfg.Enabled {
		return Client{
			Config: cfg,
			Hub:    nil,
		}, nil
	}

	if cfg.SampleRate == nil || len(cfg.SampleRate) == 0 {
		cfg.SampleRate = map[sentry.Level]float64{
			sentry.LevelError: 1.0,
			sentry.LevelFatal: 1.0,
		}
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn: cfg.DSN,

		ServerName:  cfg.AppName,
		Environment: cfg.Environment,
		Dist:        cfg.Dist,
		Release:     cfg.Release,

		Debug: cfg.Debug,

		// Removing sentry default integrations
		Integrations: func([]sentry.Integration) []sentry.Integration { return nil },

		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if len(event.Exception) > 1 {
				// If there's more than one cause, every cause is a frame.
				// Sentry add unneeded stacktrace for first frame of exception.
				// We are removing it because it brings nothing useful for researching.
				// But if there's only one frame with stacktrace, leave
				event.Exception[len(event.Exception)-1].Stacktrace = nil
			}
			return event
		},

		Transport: &sentry.HTTPTransport{
			BufferSize: cfg.TransportBufferSize,
			Timeout:    cfg.TransportTimeout,
		},
	})
	if err != nil {
		return Client{}, errors.Wrap(err, "failed to init sentry")
	}

	cutHeaders := map[string]bool{
		"Cookie":        true,
		"Authorization": true,
	}

	for _, header := range cfg.CutHeaders {
		cutHeaders[header] = true
	}

	return Client{
		Config:      cfg,
		Hub:         sentry.CurrentHub().Clone(),
		CutHeaders:  cutHeaders,
		defaultTags: defaultTags(),

		// nolint:gosec
		Rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (c Client) SendException(
	ctx context.Context,
	exception Exception,
) (string, error) {
	if !c.Config.Enabled {
		return "", nil
	}

	ok := c.checkIsAllowedBySampleRate(exception.Level)
	if !ok {
		return "", ErrRatelimit
	}

	hub := c.Hub.Clone()
	hub.Scope().SetRequest(c.newFilteredSentryRequest(exception.Request))
	hub.Scope().SetTags(c.defaultTags)
	hub.Scope().SetTags(map[string]string{
		"path":       exception.Path,
		"request-id": exception.RequestID,
	})
	hub.Scope().SetTags(exception.CustomTags)
	hub.Scope().SetTags(errorTags(exception.Error))
	hub.Scope().SetUser(exception.User)
	hub.Scope().SetExtras(mergeExtra(getExtra(ctx), exception.Extra))

	evt := eventFromError(exception.Error, exception.Level)
	evtID := hub.CaptureEvent(evt)
	if evt == nil {
		return "", errors.New("event id is nil after capturing exception")
	}

	return string(*evtID), nil
}

var ErrRatelimit = errors.Typed("sentry_ratelimit", "sentry event ratelimited")

func (c Client) checkIsAllowedBySampleRate(level Level) bool {
	levelRate := c.Config.SampleRate[level]
	return levelRate > 0 && levelRate > c.Rand.Float64()
}

func (c Client) newFilteredSentryRequest(req *http.Request) *http.Request {
	if req == nil {
		return nil
	}

	sentryReq := *req
	sentryReq.Header = req.Header.Clone()
	sentryReq.Body = nil

	for key := range c.CutHeaders {
		sentryReq.Header.Del(key)
	}

	return &sentryReq
}

type extraKey struct{}

func InitContextExtra(ctx context.Context) context.Context {
	return context.WithValue(ctx, extraKey{}, map[string]interface{}{})
}

func AddContextExtra(ctx context.Context, key string, value interface{}) {
	extra := getExtra(ctx)
	if extra != nil {
		extra[key] = value
	}
}

func getExtra(ctx context.Context) map[string]interface{} {
	extraRaw := ctx.Value(extraKey{})
	if extraRaw == nil {
		return nil
	}

	return extraRaw.(map[string]interface{})
}

func mergeExtra(extra1, extra2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(extra1)+len(extra2))
	for k, v := range extra1 {
		result[k] = v
	}
	for k, v := range extra2 {
		result[k] = v
	}

	return result
}

// eventFromError done according to original function eventFromException, called by CaptureException
func eventFromError(err error, level sentry.Level) *sentry.Event {
	if level == "" {
		level = LevelError
	}

	return &sentry.Event{
		Level: level,
		Exception: []sentry.Exception{{
			Value:      err.Error(),
			Type:       reflect.TypeOf(err).String(),
			Stacktrace: stacktraceFromError(err),
		}},
	}
}

func stacktraceFromError(err error) *sentry.Stacktrace {
	tracer := errors.ExtractDeepestStacktracer(err)
	if tracer == nil {
		return sentry.NewStacktrace()
	}

	stack := sentry.ExtractStacktrace(tracer)
	if stack == nil {
		return sentry.NewStacktrace()
	}

	return stack
}

func errorTags(err error) map[string]string {
	return map[string]string{
		"error_stack": strings.Join(errors.TypeStack(err), " -> "),
	}
}
