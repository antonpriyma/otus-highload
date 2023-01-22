package combinedstat

import "github.com/antonpriyma/otus-highload/pkg/stat"

type registry struct {
	Registries []stat.Registry
}

func NewRegistry(registries ...stat.Registry) stat.Registry {
	return registry{Registries: registries}
}

func (r registry) ForSubsystem(subsystem string) stat.Registry {
	newRegistries := make([]stat.Registry, 0, len(r.Registries))
	for _, r := range r.Registries {
		newRegistries = append(newRegistries, r.ForSubsystem(subsystem))
	}

	return registry{Registries: newRegistries}
}
