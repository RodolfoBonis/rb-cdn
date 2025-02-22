// http-error_test.go
package errors

import (
	"net/http"
	"testing"

	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/stretchr/testify/assert"
)

func TestHttpError(t *testing.T) {
	// Save original function and restore after tests
	originalGetEnv := getEnvironment
	defer func() { getEnvironment = originalGetEnv }()

	t.Run("should create new HTTP error in development environment", func(t *testing.T) {
		getEnvironment = func() string { return entities.Environment.Development }

		httpErr := NewHTTPError(http.StatusBadRequest, "test error")

		assert.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
		assert.Equal(t, "test error", httpErr.Message)
		assert.NotEmpty(t, httpErr.StackTrace)
	})

	t.Run("should create new HTTP error in production environment", func(t *testing.T) {
		getEnvironment = func() string { return entities.Environment.Production }

		httpErr := NewHTTPError(http.StatusBadRequest, "test error")

		assert.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
		assert.Equal(t, "test error", httpErr.Message)
		assert.Empty(t, httpErr.StackTrace)
	})

	t.Run("should convert HTTP error to map", func(t *testing.T) {
		getEnvironment = func() string { return entities.Environment.Development }

		httpErr := NewHTTPError(http.StatusBadRequest, "test error")
		errMap := httpErr.ToMap()

		assert.Equal(t, http.StatusBadRequest, errMap["code"])
		assert.Equal(t, "test error", errMap["message"])
		assert.NotEmpty(t, errMap["stack_trace"])
	})

	t.Run("should handle different status codes", func(t *testing.T) {
		getEnvironment = func() string { return entities.Environment.Production }

		testCases := []struct {
			name       string
			statusCode int
			message    string
		}{
			{"not found", http.StatusNotFound, "not found"},
			{"unauthorized", http.StatusUnauthorized, "unauthorized"},
			{"server error", http.StatusInternalServerError, "server error"},
			{"bad request", http.StatusBadRequest, "bad request"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				httpErr := NewHTTPError(tc.statusCode, tc.message)
				assert.Equal(t, tc.statusCode, httpErr.StatusCode)
				assert.Equal(t, tc.message, httpErr.Message)
			})
		}
	})
}
