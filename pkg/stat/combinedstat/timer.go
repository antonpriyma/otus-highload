package combinedstat

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
	ctors := make([]stat.TimerCtor, 0, len(r.Registries))
	for _, reg := range r.Registries {
		ctors = append(ctors, reg.Timer(name, buckets, labelNames, constLabels))
	}

	return timerCtor{Ctors: ctors}
}

type timer struct {
	Timers []stat.Timer
}

func (t timer) WithLabels(labels stat.Labels) stat.Timer {
	for i, tt := range t.Timers {
		t.Timers[i] = tt.WithLabels(labels)
	}

	return t
}

func (t timer) WithCaller() stat.Timer {
	return t.WithLabels(stat.CallerLabel())
}

func (t timer) Start() stat.Timer {
	for i, tt := range t.Timers {
		t.Timers[i] = tt.Start()
	}

	return t
}

func (t timer) Stop() {
	for _, tt := range t.Timers {
		tt.Stop()
	}
}

type timerCtor struct {
	Ctors []stat.TimerCtor
}

func (tc timerCtor) Timer(ctx context.Context) stat.Timer {
	timers := make([]stat.Timer, 0, len(tc.Ctors))
	for _, ctor := range tc.Ctors {
		timers = append(timers, ctor.Timer(ctx))
	}

	return timer{Timers: timers}
}
