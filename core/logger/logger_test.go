package logger

import (
	"encoding/json"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"os"
	"testing"
)

func setupTestLogger() (*CustomLogger, *observer.ObservedLogs) {
	core, recorded := observer.New(zap.InfoLevel)
	testLogger := zap.New(core)
	return &CustomLogger{logger: testLogger}, recorded
}

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name        string
		environment string
	}{
		{
			name:        "Development Environment",
			environment: entities.Environment.Development,
		},
		{
			name:        "Production Environment",
			environment: entities.Environment.Production,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("ENV", tt.environment)
			if err != nil {
				return
			}

			err = os.Setenv("NEW_RELIC_LICENSE_KEY", "40charslicensekeyrequiredfortestingpurpo")
			if err != nil {
				return
			}
			InitLogger()
			assert.NotNil(t, Log)
			assert.NotNil(t, Log.logger)
		})
	}
}

func TestCustomLogger_Info(t *testing.T) {
	logger, recorded := setupTestLogger()

	testCases := []struct {
		name     string
		message  string
		jsonData map[string]interface{}
	}{
		{
			name:    "Info without JSON",
			message: "test info message",
		},
		{
			name:    "Info with JSON",
			message: "test info with json",
			jsonData: map[string]interface{}{
				"key": "value",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.jsonData != nil {
				logger.Info(tc.message, tc.jsonData)
			} else {
				logger.Info(tc.message)
			}

			logs := recorded.All()
			assert.NotEmpty(t, logs)
			assert.Equal(t, tc.message, logs[len(logs)-1].Message)
			assert.Equal(t, zap.InfoLevel, logs[len(logs)-1].Level)
		})
	}
}

func TestCustomLogger_Warning(t *testing.T) {
	logger, recorded := setupTestLogger()

	testCases := []struct {
		name     string
		message  string
		jsonData map[string]interface{}
	}{
		{
			name:    "Warning without JSON",
			message: "test warning message",
		},
		{
			name:    "Warning with JSON",
			message: "test warning with json",
			jsonData: map[string]interface{}{
				"key": "value",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.jsonData != nil {
				logger.Warning(tc.message, tc.jsonData)
			} else {
				logger.Warning(tc.message)
			}

			logs := recorded.All()
			assert.NotEmpty(t, logs)
			assert.Equal(t, tc.message, logs[len(logs)-1].Message)
			assert.Equal(t, zap.WarnLevel, logs[len(logs)-1].Level)
		})
	}
}

func TestCustomLogger_Error(t *testing.T) {
	logger, recorded := setupTestLogger()

	testCases := []struct {
		name     string
		message  string
		jsonData map[string]interface{}
	}{
		{
			name:    "Error without JSON",
			message: "test error message",
		},
		{
			name:    "Error with JSON",
			message: "test error with json",
			jsonData: map[string]interface{}{
				"key": "value",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.jsonData != nil {
				logger.Error(tc.message, tc.jsonData)
			} else {
				logger.Error(tc.message)
			}

			logs := recorded.All()
			assert.NotEmpty(t, logs)
			assert.Equal(t, tc.message, logs[len(logs)-1].Message)
			assert.Equal(t, zap.ErrorLevel, logs[len(logs)-1].Level)
		})
	}
}

func TestLogData_JSON(t *testing.T) {
	logData := LogData{
		Level:   "info",
		Message: "test message",
		JSON: map[string]interface{}{
			"key": "value",
		},
	}

	jsonBytes, err := json.Marshal(logData)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)

	var unmarshalled LogData
	err = json.Unmarshal(jsonBytes, &unmarshalled)
	assert.NoError(t, err)
	assert.Equal(t, logData.Level, unmarshalled.Level)
	assert.Equal(t, logData.Message, unmarshalled.Message)
	assert.Equal(t, logData.JSON["key"], unmarshalled.JSON["key"])
}
