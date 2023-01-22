package loggerstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Counter(name string, _ []string, constLabels stat.Labels) stat.CounterCtor {
	return counterCtor{
		Name:        r.fullname(name),
		ConstLabels: constLabels,
		Logger:      r.Logger,
	}
}

type counter struct {
	metric
}

func (c *counter) WithLabels(labels stat.Labels) stat.Counter {
	c.Labels.Extend(labels)

	return c
}

func (c *counter) WithCaller() stat.Counter {
	c.Labels.Extend(stat.CallerLabel())

	return c
}

func (c *counter) Add(val float64) {
	c.Observe(val)
}

type counterCtor struct {
	Name        string
	ConstLabels stat.Labels
	Logger      log.Logger
}

func (cc counterCtor) Counter(ctx context.Context) stat.Counter {
	return &counter{
		metric: metric{
			Name:        cc.Name,
			Labels:      stat.Labels{},
			ConstLabels: cc.ConstLabels,
			Storage:     contextStorage(ctx, cc.Name, cc.Logger),
		},
	}
}
