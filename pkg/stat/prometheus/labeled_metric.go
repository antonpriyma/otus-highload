package prometheus

import (
	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/prometheus/client_golang/prometheus"
)

type labeledMetric struct {
	DefinedLabelNames []string
	ProvidedLabels    stat.Labels
}

func (l labeledMetric) PromLabels() prometheus.Labels {
	return prometheus.Labels(l.ProvidedLabels.ForKeys(l.DefinedLabelNames))
}

func (l labeledMetric) ProvideLabels(labels stat.Labels) {
	l.ProvidedLabels.Extend(labels)
}
