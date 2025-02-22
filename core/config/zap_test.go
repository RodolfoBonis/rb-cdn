package config

import (
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

// Mock New Relic application for testing
type mockNewRelicApp struct {
	*newrelic.Application
}

func TestZapConfig(t *testing.T) {
	tests := []struct {
		name    string
		app     *newrelic.Application
		wantErr bool
	}{
		{
			name:    "successful initialization",
			app:     &newrelic.Application{},
			wantErr: false,
		},
		{
			name:    "nil application",
			app:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("ZapConfig() panic = %v, wantErr %v", r, tt.wantErr)
					}
				}
			}()

			logger := ZapConfig(tt.app)

			if !tt.wantErr {
				assert.NotNil(t, logger)
				assert.IsType(t, &zap.Logger{}, logger)
			}
		})
	}
}

func TestZapTestConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful initialization",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("ZapTestConfig() panic = %v, wantErr %v", r, tt.wantErr)
					}
				}
			}()

			logger := ZapTestConfig()

			assert.NotNil(t, logger)
			assert.IsType(t, &zap.Logger{}, logger)

			// Test logging functionality
			logger.Info("test message")
			logger.Error("test error")
		})
	}
}
