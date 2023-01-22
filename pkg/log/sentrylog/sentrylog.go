package sentrylog

import (
	"context"
	"fmt"

	"github.com/antonpriyma/otus-highload/pkg/clients/sentry"
	"github.com/antonpriyma/otus-highload/pkg/debug"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

type Resolver func(ctx context.Context) sentry.Exception

type logger struct {
	log.Logger
	Sentry   sentry.Client
	Err      error
	Ctx      context.Context
	Resolver Resolver
}

func NewSentryLogger(ctx context.Context, log log.Logger, client sentry.Client, resolver Resolver) log.Logger {
	return logger{
		Logger:   log,
		Sentry:   client,
		Resolver: resolver,
	}.ForCtx(ctx)
}

func (l logger) ForCtx(ctx context.Context) log.Logger {
	l.Logger = l.Logger.ForCtx(ctx)
	l.Ctx = ctx
	return l
}

func (l logger) WithField(key string, val interface{}) log.Logger {
	l.Logger = l.Logger.WithField(key, val)
	return l
}

func (l logger) WithFields(fields log.Fields) log.Logger {
	l.Logger = l.Logger.WithFields(fields)
	return l
}

func (l logger) WithError(err error) log.Logger {
	l.Logger = l.Logger.WithError(err)
	l.Err = err
	return l
}

func (l logger) Warn(args ...interface{}) {
	l = l.send(sentry.LevelWarning, args...)
	l.Logger.Warn(args...)
}

func (l logger) Warnf(format string, args ...interface{}) {
	l = l.sendf(sentry.LevelWarning, format, args...)
	l.Logger.Warnf(format, args...)
}

func (l logger) Error(args ...interface{}) {
	l = l.send(sentry.LevelError, args...)
	l.Logger.Error(args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l = l.sendf(sentry.LevelError, format, args...)
	l.Logger.Errorf(format, args...)
}

func (l logger) Fatal(args ...interface{}) {
	l = l.send(sentry.LevelFatal, args...)
	l.Logger.Fatal(args...)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l = l.sendf(sentry.LevelFatal, format, args...)
	l.Logger.Fatalf(format, args...)
}

func (l logger) send(level sentry.Level, args ...interface{}) logger {
	if l.Err == nil {
		return l
	}

	sentryID := l.sendSentryEvent(l.Ctx, level, errors.Wrap(l.Err, fmt.Sprint(args...)))
	return l.WithField("sentry-id", sentryID).(logger)
}

func (l logger) sendf(level sentry.Level, format string, args ...interface{}) logger {
	if l.Err == nil {
		return l
	}

	sentryID := l.sendSentryEvent(l.Ctx, level, errors.Wrapf(l.Err, format, args...))
	return l.WithField("sentry-id", sentryID).(logger)
}

func (l logger) sendSentryEvent(ctx context.Context, level sentry.Level, err error) string {
	defer func() {
		r := errors.RecoverError(recover())
		if r != nil {
			l.Logger.ForCtx(ctx).
				WithError(r).
				WithField("stack", debug.ErrorStackTrace(r)).
				Error("panic during sentrylog processing")
		}
	}()

	if ctx == nil {
		ctx = context.Background()
		l.Logger.ForCtx(ctx).
			WithField("stack", debug.StackTrace()).
			Error("sentrylog empty context")
	}

	if ctx.Err() != nil {
		return "ctx cancelled"
	}

	exception := l.Resolver(ctx)
	exception.Error = err
	exception.Level = level

	sentryID, err := l.Sentry.SendException(ctx, exception)
	if errors.Is(err, sentry.ErrRatelimit) {
		return "ratelimited"
	}
	if err != nil {
		l.Logger.ForCtx(ctx).WithError(err).Error("failed to send sentry event")
		return "-"
	}

	return sentryID
}
