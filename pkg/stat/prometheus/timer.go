package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/prometheus/client_golang/prometheus"
)

type timer struct {
	Obs       stat.Histogram
	StartTime time.Time
}

func (r registry) Timer(
	name string,
	buckets []float64,
	labelNames []string,
	constLabels stat.Labels,
) stat.TimerCtor {
	return timerCtor{
		LabelNames: labelNames,
		Vec: r.Provider.HistogramVec(
			r.AppName,
			r.Subsystem,
			fmt.Sprintf("%s_seconds", name),
			buckets,
			prometheus.Labels(constLabels),
			labelNames,
		),
	}
}

func (t *timer) WithLabels(labels stat.Labels) stat.Timer {
	t.Obs.WithLabels(labels)

	return t
}

func (t *timer) WithCaller() stat.Timer {
	t.Obs.WithLabels(stat.CallerLabel())

	return t
}

func (t *timer) Start() stat.Timer {
	t.StartTime = time.Now()
	return t
}

func (t *timer) Stop() {
	t.Obs.Observe(time.Since(t.StartTime).Seconds())
}

type timerCtor struct {
	LabelNames []string
	Vec        prometheus.ObserverVec
}

func (tc timerCtor) Timer(_ context.Context) stat.Timer {
	return &timer{
		Obs: &histogram{
			labeledMetric: labeledMetric{
				DefinedLabelNames: tc.LabelNames,
				ProvidedLabels:    make(map[string]string, len(tc.LabelNames)),
			},
			Vec: tc.Vec,
		},
	}
}
