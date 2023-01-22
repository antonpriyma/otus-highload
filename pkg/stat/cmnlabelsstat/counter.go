package cmnlabelsstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Counter(name string, labelNames []string, constLabels stat.Labels) stat.CounterCtor {
	return counterCtor{
		cc: r.Registry.Counter(name, append(labelNames, r.CommonLabelNames...), constLabels),
	}
}

type counterCtor struct {
	cc stat.CounterCtor
}

func (cc counterCtor) Counter(ctx context.Context) stat.Counter {
	return cc.cc.Counter(ctx).WithLabels(getCommonLabels(ctx))
}
