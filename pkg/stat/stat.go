package stat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/utils"
)

const callerSkip = 2

type Registry interface {
	ForSubsystem(subsystem string) Registry
	Counter(name string, labelNames []string, constLabels Labels) CounterCtor
	Gauge(name string, labelNames []string, constLabels Labels) GaugeCtor
	Histogram(name string, buckets []float64, labelNames []string, constLabels Labels) HistogramCtor
	Timer(name string, buckets []float64, labelNames []string, constLabels Labels) TimerCtor
	Summary(name string, quantiles []float64, labelNames []string, constLabels Labels) SummaryCtor
}

type CounterCtor interface {
	Counter(ctx context.Context) Counter
}

type GaugeCtor interface {
	Gauge(ctx context.Context) Gauge
}

type HistogramCtor interface {
	Histogram(ctx context.Context) Histogram
}

type SummaryCtor interface {
	Summary(ctx context.Context) Summary
}

type TimerCtor interface {
	Timer(ctx context.Context) Timer
}

type Counter interface {
	WithLabels(labels Labels) Counter
	WithCaller() Counter
	Add(val float64)
}

type Gauge interface {
	WithLabels(labels Labels) Gauge
	WithCaller() Gauge
	Add(val float64)
	Sub(val float64)
}

type Summary interface {
	WithLabels(labels Labels) Summary
	WithCaller() Summary
	Observe(val float64)
}

type Histogram interface {
	WithLabels(labels Labels) Histogram
	WithCaller() Histogram
	Observe(val float64)
}

type Timer interface {
	WithLabels(labels Labels) Timer
	WithCaller() Timer
	Start() Timer
	Stop()
}

type Labels map[string]string

func (l Labels) Extend(l2 Labels) {
	for k, v := range l2 {
		l[k] = v
	}
}

func (l Labels) ForKeys(keys []string) Labels {
	result := make(Labels, len(keys))
	for _, key := range keys {
		val, ok := l[key]
		if ok {
			result[key] = val
		} else {
			result[key] = "unknown"
		}
	}

	return result
}

func MergeLabels(l1, l2 Labels) Labels {
	res := make(Labels, len(l1)+len(l2))
	res.Extend(l1)
	res.Extend(l2)
	return res
}

func CallerLabel() Labels {
	return Labels{"method": utils.Caller(callerSkip)}
}
