package config

import (
	"fmt"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/getsentry/sentry-go"
	"os"
)

// Move initialization function to package level for testing
var sentryInit = sentry.Init

func SentryConfig() {
	if os.Getenv("ENVIRONMENT") == entities.Environment.Test {
		return
	}

	if err := sentryInit(sentry.ClientOptions{
		Dsn:              EnvSentryDSN(),
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
		osExit(1)
	}
}
