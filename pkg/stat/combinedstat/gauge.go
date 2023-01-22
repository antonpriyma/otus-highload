package combinedstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Gauge(name string, labelNames []string, constLabels stat.Labels) stat.GaugeCtor {
	ctors := make([]stat.GaugeCtor, 0, len(r.Registries))
	for _, reg := range r.Registries {
		ctors = append(ctors, reg.Gauge(name, labelNames, constLabels))
	}

	return gaugeCtor{ctors: ctors}
}

type gauge struct {
	gauges []stat.Gauge
}

func (g gauge) WithLabels(labels stat.Labels) stat.Gauge {
	for i, gg := range g.gauges {
		g.gauges[i] = gg.WithLabels(labels)
	}

	return g
}

func (g gauge) WithCaller() stat.Gauge {
	return g.WithLabels(stat.CallerLabel())
}

func (g gauge) Add(val float64) {
	for _, gg := range g.gauges {
		gg.Add(val)
	}
}

func (g gauge) Sub(val float64) {
	for _, gg := range g.gauges {
		gg.Sub(val)
	}
}

type gaugeCtor struct {
	ctors []stat.GaugeCtor
}

func (gc gaugeCtor) Gauge(ctx context.Context) stat.Gauge {
	gauges := make([]stat.Gauge, 0, len(gc.ctors))
	for _, ctor := range gc.ctors {
		gauges = append(gauges, ctor.Gauge(ctx))
	}

	return gauge{gauges: gauges}
}
