package combinedstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Counter(name string, labelNames []string, constLabels stat.Labels) stat.CounterCtor {
	ctors := make([]stat.CounterCtor, 0, len(r.Registries))
	for _, reg := range r.Registries {
		ctors = append(ctors, reg.Counter(name, labelNames, constLabels))
	}

	return counterCtor{ctors: ctors}
}

type counter struct {
	counters []stat.Counter
}

func (c counter) WithLabels(labels stat.Labels) stat.Counter {
	for i, counter := range c.counters {
		c.counters[i] = counter.WithLabels(labels)
	}

	return c
}

func (c counter) WithCaller() stat.Counter {
	return c.WithLabels(stat.CallerLabel())
}

func (c counter) Add(val float64) {
	for _, counter := range c.counters {
		counter.Add(val)
	}
}

type counterCtor struct {
	ctors []stat.CounterCtor
}

func (cc counterCtor) Counter(ctx context.Context) stat.Counter {
	counters := make([]stat.Counter, 0, len(cc.ctors))

	for _, ctor := range cc.ctors {
		counters = append(counters, ctor.Counter(ctx))
	}

	return counter{counters: counters}
}
