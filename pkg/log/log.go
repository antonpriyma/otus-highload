package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	App   string `mapstructure:"app"`
	Level Level  `mapstructure:"level"`
}

func (c Config) Validate() error {
	if c.App == "" {
		return errors.New("empty app")
	}

	if _, ok := configLevels[c.Level]; !ok {
		return errors.Errorf("unknown log level: %q", c.Level)
	}

	return nil
}

func NewLogrusLogger(cfg Config) (Logger, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logrusLogger := logrus.New()
	logrusLogger.SetLevel(configLevels[cfg.Level])
	logrusLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	l := logrusLogger.WithField("app", cfg.App)

	hostname, err := os.Hostname()
	if err == nil {
		l = l.WithField("host", hostname)
	}

	return logger{Entry: l}, nil
}

var def Logger

func init() {
	var err error
	def, err = NewLogrusLogger(Config{
		App:   "initializing",
		Level: "debug",
	})
	if err != nil {
		panic(fmt.Sprintf("failed to init default logger: %s", err))
	}
}

func Default() Logger {
	return def
}

type logger struct {
	*logrus.Entry
}

func (l logger) ForCtx(ctx context.Context) Logger {
	if isDebugLoggingEnabled(ctx) {
		l.Entry.Level = logrus.DebugLevel
	}

	return l.WithFields(GetCtxFields(ctx))
}

func (l logger) WithFields(fields Fields) Logger {
	if fields == nil {
		return l
	}

	return logger{Entry: l.Entry.WithFields(logrus.Fields(fields))}
}

func (l logger) WithField(key string, val interface{}) Logger {
	return logger{Entry: l.Entry.WithField(key, val)}
}

func (l logger) WithError(err error) Logger {
	return logger{Entry: l.Entry.WithError(err)}
}

func (l logger) Debug(args ...interface{}) { l.Debugln(args...) }
func (l logger) Info(args ...interface{})  { l.Infoln(args...) }
func (l logger) Warn(args ...interface{})  { l.Warnln(args...) }
func (l logger) Error(args ...interface{}) { l.Errorln(args...) }
func (l logger) Fatal(args ...interface{}) { l.Fatalln(args...) }
func (l logger) Print(args ...interface{}) { l.Println(args...) }

func (l logger) Log(level Level, args ...interface{}) { l.Logln(configLevels[level], args...) }

func (l logger) Logf(level Level, format string, args ...interface{}) {
	l.Entry.Logf(configLevels[level], format, args...)
}
