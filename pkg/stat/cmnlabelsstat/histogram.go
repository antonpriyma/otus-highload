package cmnlabelsstat

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
	return histogramCtor{
		hc: r.Registry.Histogram(name, buckets, append(labelNames, r.CommonLabelNames...), constLabels),
	}
}

type histogramCtor struct {
	hc stat.HistogramCtor
}

func (hc histogramCtor) Histogram(ctx context.Context) stat.Histogram {
	return hc.hc.Histogram(ctx).WithLabels(getCommonLabels(ctx))
}
