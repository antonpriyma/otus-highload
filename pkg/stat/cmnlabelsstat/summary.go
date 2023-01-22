package cmnlabelsstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Summary(
	name string,
	quantiles []float64,
	labelNames []string,
	constLabels stat.Labels,
) stat.SummaryCtor {
	return summaryCtor{
		sc: r.Registry.Summary(name, quantiles, append(labelNames, r.CommonLabelNames...), constLabels),
	}
}

type summaryCtor struct {
	sc stat.SummaryCtor
}

func (sc summaryCtor) Summary(ctx context.Context) stat.Summary {
	return sc.sc.Summary(ctx).WithLabels(getCommonLabels(ctx))
}
