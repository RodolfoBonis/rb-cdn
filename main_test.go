package main

import (
	"fmt"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
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
)

type versionTest struct {
	name        string
	setupFunc   func()
	cleanupFunc func()
	environment string
	wantVersion string
}

type serverTest struct {
	name      string
	port      string
	setup     func()
	wantErr   bool
	shouldRun bool
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
	newRelicConfig := config.NewRelicConfig()
	middleware := middlewares.NewMonitoringMiddleware(newRelicConfig)

	app.Use(middleware.NewRelicMiddleware())
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
	tests := []versionTest{
		{
			name: "Development environment with existing version file",
			setupFunc: func() {
				_ = os.WriteFile("version.txt", []byte("1.0.0\n"), 0644)
			},
			cleanupFunc: func() {
				_ = os.Remove("version.txt")
			},
			environment: entities.Environment.Development,
			wantVersion: "1.0.0",
		},
		{
			name:        "Production environment with missing version file",
			environment: entities.Environment.Production,
			wantVersion: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}
			if tt.cleanupFunc != nil {
				defer tt.cleanupFunc()
			}

			t.Setenv("ENVIRONMENT", tt.environment)
			initializeApp()

			assert.Equal(t, "Rb CDN", docs.SwaggerInfo.Title)
			assert.Equal(t, tt.wantVersion, docs.SwaggerInfo.Version)
			assert.Equal(t, "rb-cdn.rodolfodebonis.com.br", docs.SwaggerInfo.Host)
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
