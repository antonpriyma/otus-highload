package loggerstat

import "context"

type dummyStorage struct{}

func (s dummyStorage) StoreMetric(m metricVal) {}

func (s dummyStorage) Metrics() map[string]aggregatedMetric {
	return map[string]aggregatedMetric{}
}

func InitDummyForCtx(ctx context.Context) context.Context {
	return initStorageForCtx(ctx, dummyStorage{})
}
