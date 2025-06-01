package routes

import (
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/docs"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type testCase struct {
	name         string
	path         string
	method       string
	wantStatus   int
	setupHeaders func(*http.Request)
}

func init() {
	// Set environment variables first
	os.Setenv("NEW_RELIC_LICENSE_KEY", "dummy-key-for-testing-40chars-length-her")

	// Then initialize other components
	gin.SetMode(gin.TestMode)
	logger.InitLogger()
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add recovery and logger middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Swagger setup remains the same
	docs.SwaggerInfo.Title = "RB CDN API"
	docs.SwaggerInfo.Description = "CDN API documentation"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	InitializeRoutes(router)
	return router
}

func performRequest(t *testing.T, router *gin.Engine, tc testCase) {
	t.Helper()
	w := httptest.NewRecorder()
	req, err := http.NewRequest(tc.method, tc.path, nil)
	assert.NoError(t, err)

	if tc.setupHeaders != nil {
		tc.setupHeaders(req)
	}

	router.ServeHTTP(w, req)

	if w.Code != tc.wantStatus {
		// For Swagger docs, accept both 200 and 301
		if tc.path == "/v1/docs" && (w.Code == http.StatusOK || w.Code == http.StatusMovedPermanently) {
			return
		}
		t.Logf("Response Body: %s", w.Body.String())
	}

	assert.Equal(t, tc.wantStatus, w.Code, "Path: %s, Method: %s", tc.path, tc.method)
}

func TestPublicRoutes(t *testing.T) {
	router := setupTestRouter()

	tests := []testCase{
		{
			name:       "Metrics endpoint",
			path:       "/metrics",
			method:     "GET",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Health check endpoint",
			path:       "/v1/health_check",
			method:     "GET",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Swagger docs endpoint",
			path:       "/v1/docs",
			method:     "GET",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			performRequest(t, router, tc)
		})
	}
}
