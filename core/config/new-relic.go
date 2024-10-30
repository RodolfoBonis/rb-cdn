package config

import (
	"fmt"
	"github.com/newrelic/go-agent/v3/newrelic"
	"os"
)

func NewRelicConfig() *newrelic.Application {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(EnvNewRelic().AppName),
		newrelic.ConfigLicense(EnvNewRelic().License),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	if err != nil {
		fmt.Printf("Error Relic initialization failed: %v\n", err)
		os.Exit(1)
	}

	return app
}
