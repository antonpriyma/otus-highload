package cmnlabelsstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Gauge(name string, labelNames []string, constLabels stat.Labels) stat.GaugeCtor {
	return gaugeCtor{
		gc: r.Registry.Gauge(name, append(labelNames, r.CommonLabelNames...), constLabels),
	}
}

type gaugeCtor struct {
	gc stat.GaugeCtor
}

func (gc gaugeCtor) Gauge(ctx context.Context) stat.Gauge {
	return gc.gc.Gauge(ctx).WithLabels(getCommonLabels(ctx))
}
