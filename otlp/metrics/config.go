package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	ServiceName string
	Port        int
	Registerer  prometheus.Registerer
	Gatherer    prometheus.Gatherer
}

func (c Config) GetRegisterer() prometheus.Registerer {
	if c.Registerer == nil {
		return prometheus.DefaultRegisterer
	}
	return c.Registerer
}

func (c Config) GetGatherer() prometheus.Gatherer {
	if c.Registerer == nil {
		return prometheus.DefaultGatherer
	}
	return c.Gatherer
}
