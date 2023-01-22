package loggerstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Histogram(
	name string,
	_ []float64,
	_ []string,
	constLabels stat.Labels,
) stat.HistogramCtor {
	return histogramCtor{
		Name:        r.fullname(name),
		ConstLabels: constLabels,
		Logger:      r.Logger,
	}
}

type histogram struct {
	metric
}

func (h *histogram) WithLabels(labels stat.Labels) stat.Histogram {
	h.Labels.Extend(labels)

	return h
}

func (h *histogram) WithCaller() stat.Histogram {
	h.Labels.Extend(stat.CallerLabel())

	return h
}

type histogramCtor struct {
	Name        string
	ConstLabels stat.Labels
	Logger      log.Logger
}

func (hc histogramCtor) Histogram(ctx context.Context) stat.Histogram {
	return &histogram{
		metric: metric{
			Name:        hc.Name,
			Labels:      stat.Labels{},
			ConstLabels: hc.ConstLabels,
			Storage:     contextStorage(ctx, hc.Name, hc.Logger),
		},
	}
}
