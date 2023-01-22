package loggerstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Summary(
	name string,
	_ []float64,
	_ []string,
	constLabels stat.Labels,
) stat.SummaryCtor {
	return summaryCtor{
		Name:        r.fullname(name),
		ConstLabels: constLabels,
		Logger:      r.Logger,
	}
}

type summary struct {
	metric
}

func (s *summary) WithLabels(labels stat.Labels) stat.Summary {
	s.Labels.Extend(labels)

	return s
}

func (s *summary) WithCaller() stat.Summary {
	s.Labels.Extend(stat.CallerLabel())

	return s
}

type summaryCtor struct {
	Name        string
	ConstLabels stat.Labels
	Logger      log.Logger
}

func (sc summaryCtor) Summary(ctx context.Context) stat.Summary {
	return &summary{
		metric: metric{
			Name:        sc.Name,
			Labels:      stat.Labels{},
			ConstLabels: sc.ConstLabels,
			Storage:     contextStorage(ctx, sc.Name, sc.Logger),
		},
	}
}
