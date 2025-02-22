package config

import (
	"fmt"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSentryConfig(t *testing.T) {
	// Store original functions
	originalSentryInit := sentryInit
	originalExit := osExit

	// Restore original functions after tests
	defer func() {
		sentryInit = originalSentryInit
		osExit = originalExit
	}()

	tests := []struct {
		name       string
		dsn        string
		env        string
		mockError  error
		shouldInit bool
		shouldExit bool
	}{
		{
			name:       "successful initialization",
			dsn:        "https://test@sentry.io/123456",
			env:        entities.Environment.Development,
			mockError:  nil,
			shouldInit: true,
			shouldExit: false,
		},
		{
			name:       "initialization failure",
			dsn:        "invalid-dsn",
			env:        entities.Environment.Development,
			mockError:  fmt.Errorf("invalid DSN"),
			shouldInit: true,
			shouldExit: true,
		},
		{
			name:       "test environment",
			dsn:        "https://test@sentry.io/123456",
			env:        entities.Environment.Test,
			shouldInit: false,
			shouldExit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			os.Setenv("SENTRY_DSN", tt.dsn)
			os.Setenv("ENVIRONMENT", tt.env)
			defer func() {
				os.Unsetenv("SENTRY_DSN")
				os.Unsetenv("ENVIRONMENT")
			}()

			var initCalled bool
			var exitCalled bool

			// Mock sentry.Init
			sentryInit = func(options sentry.ClientOptions) error {
				initCalled = true
				assert.Equal(t, tt.dsn, options.Dsn)
				assert.True(t, options.EnableTracing)
				assert.Equal(t, 1.0, options.TracesSampleRate)
				return tt.mockError
			}

			// Mock os.Exit
			osExit = func(code int) {
				exitCalled = true
				assert.Equal(t, 1, code)
			}

			// Execute
			SentryConfig()

			// Assert
			assert.Equal(t, tt.shouldInit, initCalled, "init called mismatch")
			assert.Equal(t, tt.shouldExit, exitCalled, "exit called mismatch")
		})
	}
}
