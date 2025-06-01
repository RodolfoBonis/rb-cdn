package logger

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Custom response writer that implements gin.ResponseWriter
type testResponseWriter struct {
	*httptest.ResponseRecorder
	statusCode int
	written    bool
}

func newTestResponseWriter() *testResponseWriter {
	return &testResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		statusCode:       http.StatusOK,
	}
}

func (w *testResponseWriter) Status() int {
	return w.statusCode
}

func (w *testResponseWriter) Size() int {
	return w.ResponseRecorder.Body.Len()
}

func (w *testResponseWriter) Written() bool {
	return w.written
}

func (w *testResponseWriter) WriteString(s string) (int, error) {
	w.written = true
	return w.ResponseRecorder.WriteString(s)
}

func (w *testResponseWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.ResponseRecorder.Write(b)
}

func (w *testResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseRecorder.WriteHeader(code)
}

func (w *testResponseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.written = true
		w.ResponseRecorder.WriteHeader(w.statusCode)
	}
}

func (w *testResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func (w *testResponseWriter) CloseNotify() <-chan bool {
	return nil
}

func (w *testResponseWriter) Pusher() http.Pusher {
	return nil
}

func (w *testResponseWriter) Flush() {
	w.ResponseRecorder.Flush()
}

// Test functions remain the same as before

// Test functions remain the same
func TestHandleRequestBody(t *testing.T) {
	t.Run("should handle nil body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := HandleRequestBody(req)
		assert.Equal(t, "", body)
	})

	t.Run("should handle non-empty body", func(t *testing.T) {
		bodyContent := `{"test": "data"}`
		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(bodyContent))
		body := HandleRequestBody(req)
		assert.Equal(t, bodyContent, body)

		newBody, _ := io.ReadAll(req.Body)
		assert.Equal(t, bodyContent, string(newBody))
	})

	t.Run("should handle error during body read", func(t *testing.T) {
		// Create a reader that will return an error
		errorReader := &ErrorReader{err: fmt.Errorf("read error")}
		req := httptest.NewRequest(http.MethodPost, "/test", errorReader)

		body := HandleRequestBody(req)
		assert.Equal(t, "", body)
	})
}

// ErrorReader implements io.Reader interface and always returns an error
type ErrorReader struct {
	err error
}

func (r *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
func TestHandleResponseBody(t *testing.T) {
	t.Run("should create body log writer", func(t *testing.T) {
		w := newTestResponseWriter()
		blw := HandleResponseBody(w)

		assert.NotNil(t, blw)
		assert.NotNil(t, blw.Body)
		assert.NotNil(t, blw.ResponseWriter)
	})
}

func TestFormatRequestAndResponse(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		method       string
		requestBody  string
		responseBody string
		requestId    string
		status       int
		want         string
	}{
		{
			name:         "metrics endpoint",
			url:          "/metrics",
			method:       "GET",
			requestBody:  "",
			responseBody: "",
			requestId:    "123",
			status:       200,
			want:         "[Request ID: 123], Status: [200], Method: [GET], Url: /metrics",
		},
		{
			name:         "health check endpoint",
			url:          "/v1/health_check",
			method:       "GET",
			requestBody:  "",
			responseBody: "",
			requestId:    "456",
			status:       200,
			want:         "[Request ID: 456], Status: [200], Method: [GET], Url: /v1/health_check",
		},
		{
			name:         "regular endpoint",
			url:          "/api/test",
			method:       "POST",
			requestBody:  `{"test":"data"}`,
			responseBody: `{"result":"success"}`,
			requestId:    "789",
			status:       201,
			want:         `[Request ID: 789], Status: [201], Method: [POST], Url: /api/test Request Body: {"test":"data"} Response Body: {"result":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newTestResponseWriter()
			w.WriteHeader(tt.status)

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.requestBody))

			result := FormatRequestAndResponse(w, req, tt.responseBody, tt.requestId, tt.requestBody)
			assert.Equal(t, tt.want, result)
		})
	}
}
