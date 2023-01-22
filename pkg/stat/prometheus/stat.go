package prometheus

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/antonpriyma/otus-highload/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type registry struct {
	AppName   string
	Subsystem string

	Provider *provider
}

type provider struct {
	Registry promauto.Factory

	counterVecs   map[string]*prometheus.CounterVec
	gaugeVecs     map[string]*prometheus.GaugeVec
	histogramVecs map[string]*prometheus.HistogramVec
	summaryVecs   map[string]*prometheus.SummaryVec

	mu sync.Mutex
}

var unsupportedSymbolsRe = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

func NewRegistry(appName string) (stat.Registry, http.Handler) {
	appName = unsupportedSymbolsRe.ReplaceAllString(appName, "_")

	promRegistry := prometheus.NewRegistry()

	// standard golang process metrics such as goroutines count or cpu usage
	promRegistry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	promRegistry.MustRegister(prometheus.NewGoCollector())

	reg := registry{
		AppName: appName,
		Provider: &provider{
			Registry: promauto.With(promRegistry),

			counterVecs:   map[string]*prometheus.CounterVec{},
			gaugeVecs:     map[string]*prometheus.GaugeVec{},
			histogramVecs: map[string]*prometheus.HistogramVec{},
			summaryVecs:   map[string]*prometheus.SummaryVec{},
		},
	}

	handler := promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{})

	return reg, handler
}

func (r registry) ForSubsystem(subsystem string) stat.Registry {
	return registry{
		AppName:   r.AppName,
		Subsystem: subsystem,
		Provider:  r.Provider,
	}
}

func metricHashKey(
	appName string,
	subsystem string,
	name string,
	constLabels prometheus.Labels,
	labelNames []string,
) string {
	var metricKey strings.Builder

	utils.MustFprintf(&metricKey, "%s%s%s", appName, subsystem, name)

	keys := make([]string, 0, len(constLabels))
	for k := range constLabels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		utils.MustFprintf(&metricKey, "%s%s", k, constLabels[k])
	}

	sort.Strings(labelNames)
	for _, label := range labelNames {
		utils.MustFprintf(&metricKey, "%s", label)
	}

	return metricKey.String()
}
