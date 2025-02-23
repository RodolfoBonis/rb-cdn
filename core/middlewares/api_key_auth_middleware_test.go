package middlewares

import (
	"fmt"
	"github.com/RodolfoBonis/go_key_guardian"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// mockValidator implements APIKeyValidator for testing
type mockValidator struct {
	mockFunc func(apiKey string, serviceId string) (keyGuardian.ApiKeyData, error)
}

func (m *mockValidator) ValidateAPIKey(apiKey string, serviceId string) (keyGuardian.ApiKeyData, error) {
	return m.mockFunc(apiKey, serviceId)
}

func init() {
	os.Setenv("NEW_RELIC_LICENSE_KEY", "40charslicensekeyrequiredfortestingpurpo")
	gin.SetMode(gin.TestMode)
	logger.InitLogger()
}

func TestProtectWithApiKey(t *testing.T) {
	mockConfig := keyGuardian.ApiKeyData{
		ID: uuid.New(),
	}

	tests := []struct {
		name          string
		apiKey        string
		expectedCode  int
		expectedJSON  bool
		handlerCalled bool
		mockFunc      func(apiKey string, serviceId string) (keyGuardian.ApiKeyData, error)
	}{
		{
			name:          "Valid API Key",
			apiKey:        "valid-api-key",
			expectedCode:  http.StatusOK,
			expectedJSON:  false,
			handlerCalled: true,
			mockFunc: func(apiKey string, serviceId string) (keyGuardian.ApiKeyData, error) {
				return mockConfig, nil
			},
		},
		{
			name:          "Missing API Key Header",
			apiKey:        "",
			expectedCode:  http.StatusUnauthorized,
			expectedJSON:  false,
			handlerCalled: false,
			mockFunc:      nil,
		},
		{
			name:          "Invalid API Key",
			apiKey:        "invalid-key",
			expectedCode:  http.StatusUnauthorized,
			expectedJSON:  true,
			handlerCalled: false,
			mockFunc: func(apiKey string, serviceId string) (keyGuardian.ApiKeyData, error) {
				return keyGuardian.ApiKeyData{}, fmt.Errorf("invalid key")
			},
		},
		{
			name:          "Validation Error",
			apiKey:        "error-key",
			expectedCode:  http.StatusUnauthorized,
			expectedJSON:  true,
			handlerCalled: false,
			mockFunc: func(apiKey string, serviceId string) (keyGuardian.ApiKeyData, error) {
				return keyGuardian.ApiKeyData{}, fmt.Errorf("validation error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			var contextCopy *gin.Context
			handlerExecuted := false

			mockHandler := func(c *gin.Context) {
				handlerExecuted = true
				contextCopy = c
				c.Status(http.StatusOK)
			}

			// Override the validator if mockFunc is provided
			if tt.mockFunc != nil {
				OverrideValidatorForTest(&mockValidator{mockFunc: tt.mockFunc})
			}
			defer RestoreDefaultValidator()

			// Create test endpoint
			r.GET("/test", ProtectWithApiKey(mockHandler))

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.apiKey != "" {
				req.Header.Set(tagApiKey, tt.apiKey)
			}

			// Perform request
			r.ServeHTTP(w, req)

			// Common assertions
			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.handlerCalled, handlerExecuted)

			// Check response format
			if tt.expectedJSON {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
			}

			if tt.handlerCalled {
				assert.NotNil(t, contextCopy, "Context should not be nil when handler is called")
				value, exists := contextCopy.Get("configs")
				assert.True(t, exists, "configs should exist in context")
				assert.NotNil(t, value, "configs value should not be nil")
				assert.Equal(t, mockConfig, value, "configs should match mock data")
			}
		})
	}
}

func TestProtectWithApiKey_EmptyApiKey(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	mockHandler := func(c *gin.Context) {
		c.Status(http.StatusOK)
	}

	r.GET("/test", ProtectWithApiKey(mockHandler))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(tagApiKey, "")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProtectWithApiKey_ValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	mockHandler := func(c *gin.Context) {
		c.Status(http.StatusOK)
	}

	r.GET("/test", ProtectWithApiKey(mockHandler))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(tagApiKey, "invalid-format-key")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
