package watcher

import (
	"strings"

	"github.com/antonpriyma/otus-highload/pkg/log"
)

type WatchableObject interface {
	ReactOnKeyChanged(key, val string)
}

type Watcher struct {
	watches map[string]WatchableObject
	logger  log.Logger
}

func NewWatcher(watches map[string]WatchableObject, logger log.Logger) Watcher {
	return Watcher{watches: watches, logger: logger}
}

func (w Watcher) Watch(chanToListen <-chan struct{ Key, Value string }) {
	w.logger.Info("start events watcher")
	defer w.logger.Info("stop events watcher")

	for {
		event, ok := <-chanToListen
		if !ok {
			return
		}

		w.reactOnKeyChanged(event.Key, event.Value)
	}
}

func (w Watcher) reactOnKeyChanged(key, val string) {
	for prefix, watcher := range w.watches {
		if strings.HasPrefix(key, prefix) {
			watcher.ReactOnKeyChanged(strings.TrimPrefix(key, prefix), val)
		}
	}
}
