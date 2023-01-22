package cmnlabelsstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Timer(
	name string,
	buckets []float64,
	labelNames []string,
	constLabels stat.Labels,
) stat.TimerCtor {
	return timerCtor{
		tc: r.Registry.Timer(name, buckets, append(labelNames, r.CommonLabelNames...), constLabels),
	}
}

type timerCtor struct {
	tc stat.TimerCtor
}

func (tc timerCtor) Timer(ctx context.Context) stat.Timer {
	return tc.tc.Timer(ctx).WithLabels(getCommonLabels(ctx))
}
