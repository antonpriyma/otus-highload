// nolint: dupl
package prometheus

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/prometheus/client_golang/prometheus"
)

type histogram struct {
	labeledMetric
	Vec prometheus.ObserverVec
}

func (r registry) Histogram(
	name string,
	buckets []float64,
	labelNames []string,
	constLabels stat.Labels,
) stat.HistogramCtor {
	return histogramCtor{
		LabelNames: labelNames,
		Vec: r.Provider.HistogramVec(
			r.AppName,
			r.Subsystem,
			name,
			buckets,
			prometheus.Labels(constLabels),
			labelNames,
		),
	}
}

func (h *histogram) WithLabels(labels stat.Labels) stat.Histogram {
	h.ProvideLabels(labels)

	return h
}

func (h *histogram) WithCaller() stat.Histogram {
	h.ProvideLabels(stat.CallerLabel())

	return h
}

func (h *histogram) Observe(val float64) {
	h.Vec.With(h.PromLabels()).Observe(val)
}

type histogramCtor struct {
	LabelNames []string
	Vec        prometheus.ObserverVec
}

func (hc histogramCtor) Histogram(_ context.Context) stat.Histogram {
	return &histogram{
		labeledMetric: labeledMetric{
			DefinedLabelNames: hc.LabelNames,
			ProvidedLabels:    make(map[string]string, len(hc.LabelNames)),
		},
		Vec: hc.Vec,
	}
}

func (p *provider) HistogramVec(
	appName string,
	subsystem string,
	name string,
	buckets []float64,
	constLabels prometheus.Labels,
	labelNames []string,
) *prometheus.HistogramVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := metricHashKey(appName, subsystem, name, constLabels, labelNames)
	metric, ok := p.histogramVecs[key]
	if ok {
		return metric
	}

	metric = p.Registry.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   appName,
			Subsystem:   subsystem,
			Name:        name,
			Buckets:     buckets,
			ConstLabels: constLabels,
		},
		labelNames,
	)

	p.histogramVecs[key] = metric

	return metric
}
