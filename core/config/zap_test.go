package config

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestZapConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "successful initialization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := ZapConfig()
			assert.NotNil(t, logger)
			assert.IsType(t, &zap.Logger{}, logger)
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
