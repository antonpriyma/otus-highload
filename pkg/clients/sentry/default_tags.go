package sentry

import (
	"os"
	"path"
	"runtime"
)

func defaultTags() map[string]string {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	exec, err := os.Executable()
	if err != nil {
		exec = "unknown"
	}

	return map[string]string{
		"host": host,
		"exec": path.Base(exec),
		"go":   runtime.Version(),
	}
}
