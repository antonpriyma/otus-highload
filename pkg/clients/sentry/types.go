package sentry

import (
	"net/http"

	"github.com/getsentry/sentry-go"
)

type Exception struct {
	Error      error
	Level      Level
	Request    *http.Request
	RequestID  string
	Path       string
	User       User
	CustomTags map[string]string
	Extra      map[string]interface{}
}

type User = sentry.User

type Level = sentry.Level

const (
	LevelDebug   Level = sentry.LevelDebug
	LevelInfo    Level = sentry.LevelInfo
	LevelWarning Level = sentry.LevelWarning
	LevelError   Level = sentry.LevelError
	LevelFatal   Level = sentry.LevelFatal
)
