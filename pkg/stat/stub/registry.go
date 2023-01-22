package stub

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
)

func NewStubRegistry() stat.Registry {
	return stub{}
}

type stub struct{}
type metric struct{}
type counter struct{ metric }
type gauge struct{ metric }
type histogram struct{ metric }
type summary struct{ metric }
type timer struct{ metric }
type counterCtor struct{}
type gaugeCtor struct{}
type histogramCtor struct{}
type summaryCtor struct{}
type timerCtor struct{}

func (stub) ForSubsystem(string) stat.Registry { return stub{} }

func (metric) Observe(float64) {}
func (metric) Sub(float64)     {}
func (metric) Add(float64)     {}

func (counter) WithLabels(stat.Labels) stat.Counter                 { return counter{} }
func (counter) WithCaller() stat.Counter                            { return counter{} }
func (counterCtor) Counter(context.Context) stat.Counter            { return counter{} }
func (stub) Counter(string, []string, stat.Labels) stat.CounterCtor { return counterCtor{} }

func (gauge) WithLabels(stat.Labels) stat.Gauge                 { return gauge{} }
func (gauge) WithCaller() stat.Gauge                            { return gauge{} }
func (gaugeCtor) Gauge(context.Context) stat.Gauge              { return gauge{} }
func (stub) Gauge(string, []string, stat.Labels) stat.GaugeCtor { return gaugeCtor{} }

func (histogram) WithLabels(stat.Labels) stat.Histogram        { return histogram{} }
func (histogram) WithCaller() stat.Histogram                   { return histogram{} }
func (histogramCtor) Histogram(context.Context) stat.Histogram { return histogram{} }
func (stub) Histogram(string, []float64, []string, stat.Labels) stat.HistogramCtor {
	return histogramCtor{}
}

func (summary) WithLabels(stat.Labels) stat.Summary      { return summary{} }
func (summary) WithCaller() stat.Summary                 { return summary{} }
func (summaryCtor) Summary(context.Context) stat.Summary { return summary{} }
func (stub) Summary(string, []float64, []string, stat.Labels) stat.SummaryCtor {
	return summaryCtor{}
}

func (timer) Start() stat.Timer                    { return timer{} }
func (timer) Stop()                                {}
func (timer) WithLabels(stat.Labels) stat.Timer    { return timer{} }
func (timer) WithCaller() stat.Timer               { return timer{} }
func (timerCtor) Timer(context.Context) stat.Timer { return timer{} }
func (stub) Timer(string, []float64, []string, stat.Labels) stat.TimerCtor {
	return timerCtor{}
}
