package log

import (
	context "context"
)

var Null Logger = nullLogger{}

type nullLogger struct{}

func (nullLogger) ForCtx(ctx context.Context) Logger {
	return nullLogger{}
}

func (nullLogger) WithField(key string, val interface{}) Logger {
	return nullLogger{}
}

func (nullLogger) WithFields(fields Fields) Logger {
	return nullLogger{}
}

func (nullLogger) WithError(err error) Logger {
	return nullLogger{}
}

func (nullLogger) Debug(args ...interface{}) {}

func (nullLogger) Debugf(format string, args ...interface{}) {}

func (nullLogger) Info(args ...interface{}) {}

func (nullLogger) Infof(format string, args ...interface{}) {}

func (nullLogger) Warn(args ...interface{}) {}

func (nullLogger) Warnf(format string, args ...interface{}) {}

func (nullLogger) Error(args ...interface{}) {}

func (nullLogger) Errorf(format string, args ...interface{}) {}

func (nullLogger) Fatal(args ...interface{}) {}

func (nullLogger) Fatalf(format string, args ...interface{}) {}

func (nullLogger) Print(args ...interface{}) {}

func (nullLogger) Printf(format string, args ...interface{}) {}

func (nullLogger) Log(level Level, args ...interface{}) {}

func (nullLogger) Logf(level Level, format string, args ...interface{}) {}
