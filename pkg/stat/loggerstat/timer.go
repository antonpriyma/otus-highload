package loggerstat

import (
	"context"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func (r registry) Timer(
	name string,
	_ []float64,
	_ []string,
	constLabels stat.Labels,
) stat.TimerCtor {
	return timerCtor{
		Name:        r.fullname(name),
		ConstLabels: constLabels,
		Logger:      r.Logger,
	}
}

type timer struct {
	metric
	StartTime time.Time
}

func (t *timer) WithLabels(labels stat.Labels) stat.Timer {
	t.Labels.Extend(labels)

	return t
}

func (t *timer) WithCaller() stat.Timer {
	t.Labels.Extend(stat.CallerLabel())

	return t
}

func (t *timer) Start() stat.Timer {
	t.StartTime = time.Now()
	return t
}

func (t *timer) Stop() {
	t.Observe(time.Since(t.StartTime).Seconds())
}

type timerCtor struct {
	Name        string
	ConstLabels stat.Labels
	Logger      log.Logger
}

func (tc timerCtor) Timer(ctx context.Context) stat.Timer {
	return &timer{
		metric: metric{
			Name:        tc.Name,
			Labels:      stat.Labels{},
			ConstLabels: tc.ConstLabels,
			Storage:     contextStorage(ctx, tc.Name, tc.Logger),
		},
	}
}
