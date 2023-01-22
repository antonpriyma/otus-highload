package log

import (
	"github.com/sirupsen/logrus"
)

type Level string

const (
	LevelError Level = "error"
	LevelWarn  Level = "warn"
	LevelInfo  Level = "info"
	LevelDebug Level = "debug"
)

var configLevels = map[Level]logrus.Level{
	LevelError: logrus.ErrorLevel,
	LevelWarn:  logrus.WarnLevel,
	LevelInfo:  logrus.InfoLevel,
	LevelDebug: logrus.DebugLevel,
}
