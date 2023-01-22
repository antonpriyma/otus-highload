package loggerstat

import (
	"context"
	"fmt"

	"github.com/antonpriyma/otus-highload/pkg/debug"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
)

type registry struct {
	AppName   string
	Subsystem string

	Logger log.Logger
}

func NewRegistry(appName string, logger log.Logger) stat.Registry {
	return registry{
		AppName: appName,
		Logger:  logger,
	}
}

func (r registry) ForSubsystem(subsystem string) stat.Registry {
	return registry{
		AppName:   r.AppName,
		Subsystem: subsystem,
		Logger:    r.Logger,
	}
}

func (r registry) fullname(name string) string {
	if r.Subsystem == "" {
		return fmt.Sprintf("%s_%s", r.AppName, name)
	}

	return fmt.Sprintf("%s_%s_%s", r.AppName, r.Subsystem, name)
}

type metric struct {
	Name        string
	Labels      stat.Labels
	ConstLabels stat.Labels

	Storage storage
}

func (m metric) Observe(val float64) {
	if m.Storage == nil {
		return
	}

	m.Storage.StoreMetric(metricVal{
		Name:   m.Name,
		Labels: stat.MergeLabels(m.Labels, m.ConstLabels),
		Value:  val,
	})
}

func contextStorage(ctx context.Context, metricName string, logger log.Logger) storage {
	st, ok := ctx.Value(storageKey{}).(storage)
	if !ok {
		logger.ForCtx(ctx).WithFields(log.Fields{
			"metric_name": metricName,
			"stacktrace":  debug.StackTrace(),
		}).Warn("metrics storage wasn't init in context")
		return nil
	}

	return st
}
