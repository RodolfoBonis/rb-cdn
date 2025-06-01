package config

import (
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewRelicConfig(t *testing.T) {
	// Store original functions
	originalNewApp := newrelicNewApplication
	originalExit := osExit

	// Restore original functions after tests
	defer func() {
		newrelicNewApplication = originalNewApp
		osExit = originalExit
	}()

	tests := []struct {
		name       string
		appName    string
		licenseKey string
		wantExit   bool
		setupMock  func()
	}{
		{
			name:       "successful initialization",
			appName:    "test-app",
			licenseKey: "1234567890123456789012345678901234567890",
			wantExit:   false,
			setupMock: func() {
				newrelicNewApplication = func(options ...newrelic.ConfigOption) (*newrelic.Application, error) {
					return &newrelic.Application{}, nil
				}
			},
		},
		{
			name:       "invalid license key",
			appName:    "test-app",
			licenseKey: "invalid",
			wantExit:   true,
			setupMock: func() {
				newrelicNewApplication = func(options ...newrelic.ConfigOption) (*newrelic.Application, error) {
					return nil, newrelic.Error{Message: "license length is not 40"}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			os.Setenv("SERVICE_NAME", tt.appName)
			os.Setenv("NEW_RELIC_LICENSE_KEY", tt.licenseKey)
			defer func() {
				os.Unsetenv("SERVICE_NAME")
				os.Unsetenv("NEW_RELIC_LICENSE_KEY")
			}()

			// Setup mocks
			tt.setupMock()
			exitCalled := false
			osExit = func(code int) {
				exitCalled = true
				assert.Equal(t, 1, code)
			}

			// Execute
			app := NewRelicConfig()

			// Assert
			if tt.wantExit {
				assert.True(t, exitCalled)
				assert.Nil(t, app)
			} else {
				assert.False(t, exitCalled)
				assert.NotNil(t, app)
			}
		})
	}
}
