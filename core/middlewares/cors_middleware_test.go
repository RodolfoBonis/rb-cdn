package middlewares

import (
	"net/http"
	"net/http/httptest"
	_ "strings"
	"testing"
	
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Testa se o middleware CORS é configurado corretamente
func TestCorsConfig(t *testing.T) {

	middleware := Cors()

	r := gin.New()
	r.Use(middleware)
	req, _ := http.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// O middleware retorna 204 para preflight requests
	assert.Equal(t, http.StatusNoContent, w.Code)

	// O header de métodos permitidos deve conter "POST"
	allowedMethods := w.Header().Get("Access-Control-Allow-Methods")
	assert.Contains(t, allowedMethods, "POST")
}

// Testa se o middleware permite uma requisição GET válida
func TestCorsAllowsGetRequest(t *testing.T) {
	r := gin.New()
	r.Use(Cors())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

// Testa se o middleware permite uma requisição POST válida
func TestCorsAllowsPostRequest(t *testing.T) {
	r := gin.New()
	r.Use(Cors())
	r.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

// Testa se o middleware responde corretamente a uma preflight request (OPTIONS)
func TestCorsHandlesPreflightRequest(t *testing.T) {
	r := gin.New()
	r.Use(Cors())

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// O middleware retorna 204 para preflight requests
	assert.Equal(t, http.StatusNoContent, w.Code)

	// O header de métodos permitidos deve conter "POST"
	allowedMethods := w.Header().Get("Access-Control-Allow-Methods")
	assert.Contains(t, allowedMethods, "POST")
}

// Testa se um header permitido é aceito corretamente
func TestCorsAllowsSpecificHeader(t *testing.T) {
	r := gin.New()
	r.Use(Cors())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("X-Api-Key", "test-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}
