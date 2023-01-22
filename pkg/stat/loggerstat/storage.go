package loggerstat

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/antonpriyma/otus-highload/pkg/debug"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/antonpriyma/otus-highload/pkg/utils"
)

type storage interface {
	StoreMetric(m metricVal)
	Metrics() map[string]aggregatedMetric
}

type simpleStorage struct {
	mu      sync.Mutex
	metrics map[string]aggregatedMetric
}

type metricVal struct {
	Name   string
	Labels stat.Labels
	Value  float64
}

type aggregatedMetric struct {
	Name   string      `json:"name"`
	Total  float64     `json:"value"`
	Count  float64     `json:"count"`
	Avg    float64     `json:"avg"`
	Max    float64     `json:"max"`
	Min    float64     `json:"min"`
	Labels stat.Labels `json:"labels"`
}

func (s *simpleStorage) StoreMetric(m metricVal) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metricKey := metricHashKey(m)

	metric, ok := s.metrics[metricKey]
	if !ok {
		metric = aggregatedMetric{
			Name:   m.Name,
			Labels: m.Labels,
		}
	}

	metric.Count++
	metric.Total += m.Value

	if metric.Min == 0 {
		metric.Min = m.Value
	}

	metric.Min = utils.MinFloat(metric.Min, m.Value)
	metric.Max = utils.MaxFloat(metric.Max, m.Value)

	s.metrics[metricKey] = metric
}

func (s *simpleStorage) Metrics() map[string]aggregatedMetric {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.metrics
}

func prepareMetricsOutput(m map[string]aggregatedMetric) map[string][]aggregatedMetric {
	ret := map[string][]aggregatedMetric{}
	for _, v := range m {
		v.Avg = v.Total / v.Count
		ret[v.Name] = append(ret[v.Name], v)
	}

	return ret
}

func metricHashKey(val metricVal) string {
	var metricKey strings.Builder
	utils.MustFprintf(&metricKey, "%s", val.Name)

	keys := make([]string, 0, len(val.Labels))
	for k := range val.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		utils.MustFprintf(&metricKey, "%s%s", k, val.Labels[k])
	}

	return metricKey.String()
}

type storageKey struct{}

func initStorageForCtx(ctx context.Context, st storage) context.Context {
	return context.WithValue(ctx, storageKey{}, st)
}

func InitStatForCtx(ctx context.Context) context.Context {
	return initStorageForCtx(ctx, &simpleStorage{metrics: map[string]aggregatedMetric{}})
}

func PrintStat(ctx context.Context, logger log.Logger) {
	st, ok := ctx.Value(storageKey{}).(storage)
	if !ok {
		logger.ForCtx(ctx).WithFields(log.Fields{
			"stacktrace": debug.StackTrace(),
		}).Warn("metrics storage wasn't init in context")
		return
	}

	metrics := st.Metrics()
	if len(metrics) == 0 {
		return
	}

	logger.ForCtx(ctx).WithField("stat", log.AsJSON(prepareMetricsOutput(metrics))).Info("stat")
}
