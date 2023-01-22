//nolint:dupl
package combinedstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Histogram(
	name string,
	buckets []float64,
	labelNames []string,
	constLabels stat.Labels,
) stat.HistogramCtor {
	ctors := make([]stat.HistogramCtor, 0, len(r.Registries))
	for _, reg := range r.Registries {
		ctors = append(ctors, reg.Histogram(name, buckets, labelNames, constLabels))
	}

	return histogramCtor{ctors: ctors}
}

type histogram struct {
	histograms []stat.Histogram
}

func (h histogram) WithLabels(labels stat.Labels) stat.Histogram {
	for i, hh := range h.histograms {
		h.histograms[i] = hh.WithLabels(labels)
	}

	return h
}

func (h histogram) WithCaller() stat.Histogram {
	return h.WithLabels(stat.CallerLabel())
}

func (h histogram) Observe(val float64) {
	for _, hh := range h.histograms {
		hh.Observe(val)
	}
}

type histogramCtor struct {
	ctors []stat.HistogramCtor
}

func (hc histogramCtor) Histogram(ctx context.Context) stat.Histogram {
	histograms := make([]stat.Histogram, 0, len(hc.ctors))
	for _, ctor := range hc.ctors {
		histograms = append(histograms, ctor.Histogram(ctx))
	}

	return histogram{histograms: histograms}
}
