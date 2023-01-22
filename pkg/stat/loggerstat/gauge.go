package loggerstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Gauge(name string, _ []string, constLabels stat.Labels) stat.GaugeCtor {
	return gaugeCtor{
		Name:        r.fullname(name),
		ConstLabels: constLabels,
		Logger:      r.Logger,
	}
}

type gauge struct {
	metric
}

func (g *gauge) WithLabels(labels stat.Labels) stat.Gauge {
	g.Labels.Extend(labels)

	return g
}

func (g *gauge) WithCaller() stat.Gauge {
	g.Labels.Extend(stat.CallerLabel())

	return g
}

func (g *gauge) Add(val float64) {
	g.Observe(val)
}

func (g *gauge) Sub(val float64) {
	g.Observe(-val)
}

type gaugeCtor struct {
	Name        string
	ConstLabels stat.Labels
	Logger      log.Logger
}

func (gc gaugeCtor) Gauge(ctx context.Context) stat.Gauge {
	return &gauge{
		metric: metric{
			Name:        gc.Name,
			Labels:      stat.Labels{},
			ConstLabels: gc.ConstLabels,
			Storage:     contextStorage(ctx, gc.Name, gc.Logger),
		},
	}
}
