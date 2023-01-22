package prometheus

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/prometheus/client_golang/prometheus"
)

type gauge struct {
	labeledMetric
	Vec *prometheus.GaugeVec
}

func (r registry) Gauge(name string, labelNames []string, constLabels stat.Labels) stat.GaugeCtor {
	return gaugeCtor{
		LabelNames: labelNames,
		Vec:        r.Provider.GaugeVec(r.AppName, r.Subsystem, name, prometheus.Labels(constLabels), labelNames),
	}
}

func (g *gauge) WithLabels(labels stat.Labels) stat.Gauge {
	g.ProvideLabels(labels)

	return g
}

func (g *gauge) WithCaller() stat.Gauge {
	g.ProvideLabels(stat.CallerLabel())

	return g
}

func (g *gauge) Add(val float64) {
	g.Vec.With(g.PromLabels()).Add(val)
}

func (g *gauge) Sub(val float64) {
	g.Vec.With(g.PromLabels()).Sub(val)
}

type gaugeCtor struct {
	LabelNames []string
	Vec        *prometheus.GaugeVec
}

func (gc gaugeCtor) Gauge(_ context.Context) stat.Gauge {
	return &gauge{
		labeledMetric: labeledMetric{
			DefinedLabelNames: gc.LabelNames,
			ProvidedLabels:    make(map[string]string, len(gc.LabelNames)),
		},
		Vec: gc.Vec,
	}
}

func (p *provider) GaugeVec(
	appName string,
	subsystem string,
	name string,
	constLabels prometheus.Labels,
	labelNames []string,
) *prometheus.GaugeVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := metricHashKey(appName, subsystem, name, constLabels, labelNames)
	metric, ok := p.gaugeVecs[key]
	if ok {
		return metric
	}

	metric = p.Registry.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   appName,
			Subsystem:   subsystem,
			Name:        name,
			ConstLabels: constLabels,
		},
		labelNames,
	)

	p.gaugeVecs[key] = metric

	return metric
}
