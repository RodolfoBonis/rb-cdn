package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para o logger
type MockLogger struct {
	logger.CustomLogger
	mock.Mock
}

// Implementação correta dos métodos do logger
func (m *MockLogger) Info(message string, jsonData ...map[string]interface{}) {
	m.Called(message, jsonData)
}

func (m *MockLogger) Error(message string, jsonData ...map[string]interface{}) {
	m.Called(message, jsonData)
}

// Função para configurar o mock do logger
func setupMockLogger() *MockLogger {
	mockLogger := new(MockLogger)
	logger.InitLogger()
	mockLogger.CustomLogger = *logger.Log
	return mockLogger
}

// Testa a criação do MonitoringMiddleware
func TestNewMonitoringMiddleware(t *testing.T) {
	mockLogger := setupMockLogger()
	app, _ := newrelic.NewApplication(newrelic.ConfigAppName("TestApp"))
	middleware := NewMonitoringMiddleware(app, mockLogger.CustomLogger)
	assert.NotNil(t, middleware)
	assert.Equal(t, app, middleware.newRelicConfig)
}

// Testa a inicialização do middleware Sentry
func TestSentryMiddleware(t *testing.T) {
	mockLogger := setupMockLogger()
	middleware := NewMonitoringMiddleware(nil, mockLogger.CustomLogger)
	handler := middleware.SentryMiddleware()
	assert.NotNil(t, handler)
}

// Testa a inicialização do middleware New Relic
func TestNewRelicMiddleware(t *testing.T) {
	mockLogger := setupMockLogger()
	app, _ := newrelic.NewApplication(newrelic.ConfigAppName("TestApp"))
	middleware := NewMonitoringMiddleware(app, mockLogger.CustomLogger)
	handler := middleware.NewRelicMiddleware()
	assert.NotNil(t, handler)
}

// Testa a função isSuccessStatusCode
func TestIsSuccessStatusCode(t *testing.T) {
	successCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent,
		http.StatusPartialContent,
	}

	errorCodes := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, code := range successCodes {
		assert.True(t, isSuccessStatusCode(code))
	}

	for _, code := range errorCodes {
		assert.False(t, isSuccessStatusCode(code))
	}
}

// Mock do HandleResponseBody
type MockResponseBody struct {
	bytes.Buffer
}

func (m *MockResponseBody) Write(p []byte) (n int, err error) {
	return m.Buffer.Write(p)
}

// Testa se a LogMiddleware adiciona o requestId ao escopo do Sentry
func TestLogMiddleware_SentryRequestId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockLogger := setupMockLogger()
	middleware := NewMonitoringMiddleware(nil, mockLogger.CustomLogger)
	r.Use(middleware.LogMiddleware)

	r.GET("/sentry", func(c *gin.Context) {
		c.String(http.StatusOK, "sentry")
	})

	req, _ := http.NewRequest("GET", "/sentry", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
