//nolint:dupl
package combinedstat

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
	ctors := make([]stat.SummaryCtor, 0, len(r.Registries))
	for _, reg := range r.Registries {
		ctors = append(ctors, reg.Summary(name, quantiles, labelNames, constLabels))
	}

	return summaryCtor{Ctors: ctors}
}

type summary struct {
	summaries []stat.Summary
}

func (s summary) WithLabels(labels stat.Labels) stat.Summary {
	for i, ss := range s.summaries {
		s.summaries[i] = ss.WithLabels(labels)
	}

	return s
}

func (s summary) WithCaller() stat.Summary {
	return s.WithLabels(stat.CallerLabel())
}

func (s summary) Observe(val float64) {
	for _, ss := range s.summaries {
		ss.Observe(val)
	}
}

type summaryCtor struct {
	Ctors []stat.SummaryCtor
}

func (sc summaryCtor) Summary(ctx context.Context) stat.Summary {
	summaries := make([]stat.Summary, 0, len(sc.Ctors))
	for _, ctor := range sc.Ctors {
		summaries = append(summaries, ctor.Summary(ctx))
	}

	return summary{summaries: summaries}
}
