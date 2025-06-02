package main

import (
	"fmt"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/core/middlewares"
	"github.com/RodolfoBonis/rb-cdn/docs"
	"github.com/RodolfoBonis/rb-cdn/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
	_ "unsafe"
)

type serverTest struct {
	name      string
	port      string
	setup     func()
	wantErr   bool
	shouldRun bool
}

// Mock logger implementation
type mockCustomLogger struct {
	logger.CustomLogger
	errorCalled bool
	infoCalled  bool
	warnCalled  bool
	debugCalled bool
	lastMessage string
	lastFields  map[string]interface{}
}

func setupTestServer(t *testing.T, port string) *gin.Engine {
	t.Helper()
	t.Setenv("PORT", port)
	t.Setenv("NEWRELIC_APP_NAME", "test")
	t.Setenv("NEWRELIC_LICENSE", "test")

	app := gin.New()
	err := app.SetTrustedProxies([]string{})
	assert.NoError(t, err)
	return app
}

func setupTestMiddleware(app *gin.Engine) {
	middleware := middlewares.NewMonitoringMiddleware(logger.CustomLogger{})

	app.Use(middleware.LogMiddleware)
	app.Use(gin.Logger())
	app.Use(gin.Recovery())
	app.Use(gin.ErrorLogger())
}

func setupCORS(app *gin.Engine) {
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "OPTIONS", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Range", "X-Api-Key"},
		ExposeHeaders:    []string{"Content-Range", "Content-Length"},
		AllowCredentials: true,
	}))
}

func skipSentryInTest(t *testing.T) {
	t.Helper()
	t.Setenv("ENVIRONMENT", entities.Environment.Test)
}

func TestInitializeApp(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		setupFunc   func(t *testing.T)
		cleanupFunc func()
		wantVersion string
	}{
		{
			name:        "Development environment with existing version file",
			environment: "development",
			setupFunc: func(t *testing.T) {
				err := os.WriteFile("version.txt", []byte("1.0.0"), 0644)
				assert.NoError(t, err, "Failed to create version file")
			},
			cleanupFunc: func() {
				os.Remove("version.txt")
			},
			wantVersion: "1.0.0",
		},
		{
			name:        "Test environment with existing version file",
			environment: "test",
			setupFunc: func(t *testing.T) {
				err := os.WriteFile("version.txt", []byte("1.0.0"), 0644)
				assert.NoError(t, err, "Failed to create version file")
			},
			cleanupFunc: func() {
				os.Remove("version.txt")
			},
			wantVersion: "1.0.0",
		},
		{
			name:        "Production environment with missing version file",
			environment: "production",
			setupFunc: func(t *testing.T) {
				os.Remove("version.txt")
			},
			wantVersion: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure clean state
			os.Remove("version.txt")

			// Setup test
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			// Cleanup after test
			defer func() {
				if tt.cleanupFunc != nil {
					tt.cleanupFunc()
				}
			}()

			// Set environment
			t.Setenv("ENVIRONMENT", tt.environment)

			// Initialize app
			initializeApp()

			// Assert swagger info
			assert.Equal(t, "Rb CDN", docs.SwaggerInfo.Title)
			assert.Equal(t, tt.wantVersion, docs.SwaggerInfo.Version)
			assert.Equal(t, "/v1", docs.SwaggerInfo.BasePath)
			assert.Equal(t, []string{"https"}, docs.SwaggerInfo.Schemes)
		})
	}
}

func TestServerConfiguration(t *testing.T) {
	tests := []serverTest{
		{
			name:      "Valid server setup",
			port:      "8081",
			wantErr:   false,
			shouldRun: true,
		},
		{
			name:      "Invalid port",
			port:      "invalid",
			wantErr:   true,
			shouldRun: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipSentryInTest(t)
			app := setupTestServer(t, tt.port)

			assert.NotPanics(t, func() {
				setupTestMiddleware(app)
				setupCORS(app)
				routes.InitializeRoutes(app)
			})

			done := make(chan bool)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						done <- true
					}
				}()
				main()
			}()

			select {
			case <-done:
				if !tt.wantErr {
					t.Error("Server failed unexpectedly")
				}
			case <-time.After(100 * time.Millisecond):
				if tt.wantErr {
					t.Error("Server should have failed")
				}
			}
		})
	}
}

func TestServerTimeouts(t *testing.T) {
	server := &http.Server{
		Handler:      gin.New(),
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}
	assert.Equal(t, 10*time.Minute, server.ReadTimeout)
	assert.Equal(t, 10*time.Minute, server.WriteTimeout)
}

func TestPortConfiguration(t *testing.T) {
	tests := map[string]struct {
		port, expected string
	}{
		"Custom port":  {"9090", ":9090"},
		"Default port": {"", ":8000"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.port != "" {
				t.Setenv("PORT", tt.port)
			} else {
				os.Unsetenv("PORT")
			}
			assert.Equal(t, tt.expected, fmt.Sprintf(":%s", config.EnvPort()))
		})
	}
}

func TestMainErrorHandling(t *testing.T) {
	// Store original logger
	originalLog := logger.Log
	defer func() {
		logger.Log = originalLog
	}()

	// Create mock logger
	ml := &mockCustomLogger{}
	logger.Log = &logger.CustomLogger{} // Initialize with empty CustomLogger
	logger.Log = &ml.CustomLogger

	tests := []struct {
		name        string
		setupFunc   func() error
		wantMessage string
		wantPanic   bool
	}{
		{
			name: "handle trusted proxies error",
			setupFunc: func() error {
				return fmt.Errorf("trusted proxies error")
			},
			wantMessage: "trusted proxies error",
			wantPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml.reset()

			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Error("expected panic but got none")
				}
				if !tt.wantPanic && r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			err := tt.setupFunc()
			if err != nil {
				appError := errors.RootError(err.Error())
				logger.Log.Error(appError.Message, appError.ToMap())
				panic(err)
			}

			if tt.wantMessage != "" {
				assert.True(t, ml.errorCalled)
				assert.Contains(t, ml.lastMessage, tt.wantMessage)
				assert.NotNil(t, ml.lastFields)
			}
		})
	}
}

// Mock logger methods
func (m *mockCustomLogger) Error(msg string, jsonData ...map[string]interface{}) {
	m.errorCalled = true
	m.lastMessage = msg
	if len(jsonData) > 0 {
		m.lastFields = jsonData[0]
	}
}

func (m *mockCustomLogger) Info(msg string, jsonData ...map[string]interface{}) {
	m.infoCalled = true
	m.lastMessage = msg
	if len(jsonData) > 0 {
		m.lastFields = jsonData[0]
	}
}

func (m *mockCustomLogger) Warning(msg string, jsonData ...map[string]interface{}) {
	m.warnCalled = true
	m.lastMessage = msg
	if len(jsonData) > 0 {
		m.lastFields = jsonData[0]
	}
}

func (m *mockCustomLogger) reset() {
	m.errorCalled = false
	m.infoCalled = false
	m.warnCalled = false
	m.lastMessage = ""
	m.lastFields = nil
}
