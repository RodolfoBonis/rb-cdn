package config

import (
	"fmt"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func setupEnv(t *testing.T) func() {
	t.Helper()
	envVars := map[string]string{
		"PORT":                  "3000",
		"SERVICE_NAME":          "test-service",
		"NEW_RELIC_LICENSE_KEY": "test-license",
		"SERVICE_ID":            "test-id",
		"SENTRY_DSN":            "test-dsn",
		"MINIO_SERVER":          "test-server",
		"MINIO_ACCESS_ID":       "test-access-id",
		"MINIO_SECRET_KEY":      "test-secret-key",
		"DB_HOST":               "test-host",
		"DB_PORT":               "5433",
		"DB_USER":               "test-user",
		"DB_SECRET":             "test-password",
		"DB_NAME":               "test-db",
		"USER_AMQP":             "test-user",
		"PASSWORD_AMQP":         "test-password",
		"HOST_AMQP":             "test-host:5672",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
	}

	return func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}
}

func TestEnvironmentFunctions(t *testing.T) {
	cleanup := setupEnv(t)
	defer cleanup()

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{"EnvPort", EnvPort, "3000"},
		{"EnvServiceId", EnvServiceId, "test-id"},
		{"EnvMinioHost", EnvMinioHost, "test-server"},
		{"EnvMinioAccessId", EnvMinioAccessId, "test-access-id"},
		{"EnvMinioSecretKey", EnvMinioSecretKey, "test-secret-key"},
		{"EnvSentryDSN", EnvSentryDSN, "test-dsn"},
		{"EnvDBHost", EnvDBHost, "test-host"},
		{"EnvDBPort", EnvDBPort, "5433"},
		{"EnvDBUser", EnvDBUser, "test-user"},
		{"EnvDBPassword", EnvDBPassword, "test-password"},
		{"EnvDBName", EnvDBName, "test-db"},
		{"EnvServiceName", EnvServiceName, "test-service"},
		{"envUserAmqp", envUserAmqp, "test-user"},
		{"envPasswordAmqp", envPasswordAmqp, "test-password"},
		{"envHostAmqp", envHostAmqp, "test-host:5672"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.fn())
		})
	}

	t.Run("EnvAmqpConnection", func(t *testing.T) {
		expected := "amqp://test-user:test-password@test-host:5672/"
		assert.Equal(t, expected, EnvAmqpConnection())
	})
}

func TestDefaultValues(t *testing.T) {
	// Clear all environment variables
	envVars := []string{
		"PORT", "SERVICE_NAME", "NEW_RELIC_LICENSE_KEY", "SERVICE_ID",
		"MINIO_SERVER", "MINIO_ACCESS_ID", "MINIO_SECRET_KEY", "SENTRY_DSN",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_SECRET", "DB_NAME",
		"USER_AMQP", "PASSWORD_AMQP", "HOST_AMQP",
	}

	for _, v := range envVars {
		os.Unsetenv(v)
	}

	assert.Equal(t, "8000", EnvPort())
	assert.Equal(t, "", EnvServiceId())
	assert.Equal(t, "", EnvMinioHost())
	assert.Equal(t, "", EnvMinioAccessId())
	assert.Equal(t, "", EnvMinioSecretKey())
	assert.Equal(t, "", EnvSentryDSN())
	assert.Equal(t, "localhost", EnvDBHost())
	assert.Equal(t, "5432", EnvDBPort())
	assert.Equal(t, "", EnvDBUser())
	assert.Equal(t, "", EnvDBPassword())
	assert.Equal(t, "", EnvDBName())
	assert.Equal(t, "API", EnvServiceName())
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", EnvAmqpConnection())
}

func TestLoadEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	originalExit := osExit
	defer func() { osExit = originalExit }()

	tests := []struct {
		name       string
		env        string
		withFile   bool
		wantExit   bool
		filePrefix string
	}{
		{name: "production", env: entities.Environment.Production, withFile: false, wantExit: false},
		{name: "test env", env: entities.Environment.Test, withFile: true, wantExit: false}, // Changed: withFile to true
		{name: "dev no file", env: entities.Environment.Development, withFile: false, wantExit: true},
		{name: "dev with .env", env: entities.Environment.Development, withFile: true, filePrefix: ".env"},
		{name: "dev with .env.development", env: entities.Environment.Development, withFile: true, filePrefix: ".env.development"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(".env")
			os.Remove(".env.development")
			os.Remove(".env.test") // Added: remove test env file

			if tt.withFile {
				filename := tt.filePrefix
				if filename == "" {
					filename = fmt.Sprintf(".env.%s", tt.env) // Added: use env-specific file
				}
				err := os.WriteFile(filename, []byte("TEST_VAR=value"), 0644)
				assert.NoError(t, err)
			}

			os.Setenv("ENV", tt.env)
			exitCalled := false
			osExit = func(int) { exitCalled = true }

			LoadEnvVars()

			assert.Equal(t, tt.wantExit, exitCalled)
			if tt.withFile {
				assert.Equal(t, "value", os.Getenv("TEST_VAR"))
			}
		})
	}
}

func TestEnvironmentConfig(t *testing.T) {
	tests := []struct {
		env  string
		want string
	}{
		{entities.Environment.Development, entities.Environment.Development},
		{entities.Environment.Production, entities.Environment.Production},
		{entities.Environment.Test, entities.Environment.Test},
		{"", entities.Environment.Test},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			if tt.env != "" {
				os.Setenv("ENV", tt.env)
				defer os.Unsetenv("ENV")
			}
			assert.Equal(t, tt.want, EnvironmentConfig())
		})
	}
}
