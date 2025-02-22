package config

import (
	"fmt"
	"github.com/newrelic/go-agent/v3/newrelic"
)

var (
	newrelicNewApplication = newrelic.NewApplication
)

func NewRelicConfig() *newrelic.Application {
	app, err := newrelicNewApplication(
		newrelic.ConfigAppName(EnvNewRelic().AppName),
		newrelic.ConfigLicense(EnvNewRelic().License),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	if err != nil {
		fmt.Printf("Error Relic initialization failed: %v\n", err)
		osExit(1)
		return nil // for testing purposes
	}

	return app
}
