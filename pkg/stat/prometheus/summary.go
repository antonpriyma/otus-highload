// nolint: dupl
package prometheus

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/prometheus/client_golang/prometheus"
)

type summary struct {
	labeledMetric
	Vec prometheus.ObserverVec
}

func (r registry) Summary(
	name string,
	quantiles []float64,
	labelNames []string,
	constLabels stat.Labels,
) stat.SummaryCtor {
	objs := make(map[float64]float64, len(quantiles))
	for _, q := range quantiles {
		objs[q] = (1 - q) / 10
	}

	return summaryCtor{
		LabelNames: labelNames,
		Vec:        r.Provider.SummaryVec(r.AppName, r.Subsystem, name, objs, prometheus.Labels(constLabels), labelNames),
	}
}

func (s *summary) WithLabels(labels stat.Labels) stat.Summary {
	s.ProvideLabels(labels)

	return s
}

func (s *summary) WithCaller() stat.Summary {
	s.ProvideLabels(stat.CallerLabel())

	return s
}

func (s *summary) Observe(val float64) {
	s.Vec.With(s.PromLabels()).Observe(val)
}

type summaryCtor struct {
	LabelNames []string
	Vec        prometheus.ObserverVec
}

func (sc summaryCtor) Summary(_ context.Context) stat.Summary {
	return &summary{
		labeledMetric: labeledMetric{
			DefinedLabelNames: sc.LabelNames,
			ProvidedLabels:    make(map[string]string, len(sc.LabelNames)),
		},
		Vec: sc.Vec,
	}
}

func (p *provider) SummaryVec(
	appName string,
	subsystem string,
	name string,
	objectives map[float64]float64,
	constLabels prometheus.Labels,
	labelNames []string,
) *prometheus.SummaryVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := metricHashKey(appName, subsystem, name, constLabels, labelNames)
	metric, ok := p.summaryVecs[key]
	if ok {
		return metric
	}

	metric = p.Registry.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:   appName,
			Subsystem:   subsystem,
			Name:        name,
			Objectives:  objectives,
			ConstLabels: constLabels,
		},
		labelNames,
	)

	p.summaryVecs[key] = metric

	return metric
}
