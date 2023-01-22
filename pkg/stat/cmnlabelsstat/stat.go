package cmnlabelsstat

import "github.com/antonpriyma/otus-highload/pkg/stat"

type registry struct {
	CommonLabelNames []string
	Registry         stat.Registry
}

func NewRegistry(reg stat.Registry, commonLabelNames []string) stat.Registry {
	return registry{
		CommonLabelNames: commonLabelNames,
		Registry:         reg,
	}
}

func (r registry) ForSubsystem(subsystem string) stat.Registry {
	return registry{
		CommonLabelNames: r.CommonLabelNames,
		Registry:         r.Registry.ForSubsystem(subsystem),
	}
}
