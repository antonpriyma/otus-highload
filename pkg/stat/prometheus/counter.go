package prometheus

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/prometheus/client_golang/prometheus"
)

type counter struct {
	labeledMetric
	Vec *prometheus.CounterVec
}

func (r registry) Counter(name string, labelNames []string, constLabels stat.Labels) stat.CounterCtor {
	return counterCtor{
		LabelNames: labelNames,
		Vec:        r.Provider.CounterVec(r.AppName, r.Subsystem, name, prometheus.Labels(constLabels), labelNames),
	}
}

func (c *counter) WithLabels(labels stat.Labels) stat.Counter {
	c.ProvideLabels(labels)

	return c
}

func (c *counter) WithCaller() stat.Counter {
	c.ProvideLabels(stat.CallerLabel())

	return c
}

func (c *counter) Add(val float64) {
	c.Vec.With(c.PromLabels()).Add(val)
}

type counterCtor struct {
	LabelNames []string
	Vec        *prometheus.CounterVec
}

func (cc counterCtor) Counter(_ context.Context) stat.Counter {
	return &counter{
		labeledMetric: labeledMetric{
			DefinedLabelNames: cc.LabelNames,
			ProvidedLabels:    make(map[string]string, len(cc.LabelNames)),
		},
		Vec: cc.Vec,
	}
}

func (p *provider) CounterVec(
	appName string,
	subsystem string,
	name string,
	constLabels prometheus.Labels,
	labelNames []string,
) *prometheus.CounterVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := metricHashKey(appName, subsystem, name, constLabels, labelNames)
	metric, ok := p.counterVecs[key]
	if ok {
		return metric
	}

	metric = p.Registry.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   appName,
			Subsystem:   subsystem,
			Name:        name,
			ConstLabels: constLabels,
		},
		labelNames,
	)

	p.counterVecs[key] = metric

	return metric
}
