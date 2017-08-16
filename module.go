package main

import (
	"github.com/alittlebrighter/layer-metrics/datastores"
	"github.com/alittlebrighter/layer-metrics/io"
	"github.com/alittlebrighter/layer-metrics/jobs"
)

// Module is a manual dependency injector that allows dependencies to be prebuilt based off of a configuration.
type Module interface {
	Build(*AppConfig)
	DataStore() (datastores.DataStore, error)
	Controller() io.Controller
	Job(<-chan io.Command) jobs.Job
}

type Production struct {
	config *AppConfig
}

func (pm *Production) Build(config *AppConfig) {
	pm.config = config
}

func (pm *Production) DataStore() (datastores.DataStore, error) {
	return datastores.NewInfluxStore(pm.config.InfluxEndpoint, pm.config.PollRate)
}

func (pm *Production) Controller() io.Controller {
	return io.NewHTTPController(pm.config.PollRate)
}

func (pm *Production) Job(commands <-chan io.Command) jobs.Job {
	return jobs.NewFluxCapacitorStatus(commands, pm.config.PollRate)
}
