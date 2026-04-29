package bootstrap

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/minio/minio-go"
)

// fakeBucketLister implements bootstrap.BucketLister — the narrow
// interface SyncCapabilities depends on. Only ListBuckets is needed.
type fakeBucketLister struct {
	buckets []minio.BucketInfo
	err     *errors.AppError
}

func (f *fakeBucketLister) ListBuckets() ([]minio.BucketInfo, *errors.AppError) {
	return f.buckets, f.err
}

// recordingFatal captures the format string + first arg so tests
// can assert which abort path triggered, without aborting the
// process.
type recordingFatal struct {
	called atomic.Bool
	format string
	args   []any
}

func (r *recordingFatal) Fn() FatalFunc {
	return func(format string, args ...any) {
		r.called.Store(true)
		r.format = format
		r.args = args
	}
}

// startStubBackend wires fake Keycloak + management API for the
// rbauth provider HTTP path.
func startStubBackend(t *testing.T, syncStatus int, syncBody string) (*httptest.Server, *atomic.Int32, *atomic.Pointer[string]) {
	t.Helper()
	syncCalls := &atomic.Int32{}
	lastBody := &atomic.Pointer[string]{}
	mux := http.NewServeMux()
	mux.HandleFunc("/realms/", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "tok-" + time.Now().UTC().Format("150405.000000000"),
			"expires_in":   60,
			"token_type":   "Bearer",
		})
	})
	mux.HandleFunc("/v1/identities/sync", func(w http.ResponseWriter, r *http.Request) {
		syncCalls.Add(1)
		raw, _ := io.ReadAll(r.Body)
		s := string(raw)
		lastBody.Store(&s)
		w.WriteHeader(syncStatus)
		_, _ = w.Write([]byte(syncBody))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, syncCalls, lastBody
}

func newTestLogger() *logger.CustomLogger {
	// CustomLogger is initialised by InitLogger() reading env vars.
	// For the test we just rely on ZapTestConfig via Development.
	old := os.Getenv("ENV")
	_ = os.Setenv("ENV", "development")
	t := old
	defer func() { _ = os.Setenv("ENV", t) }()
	logger.InitLogger()
	return logger.Log
}

func setBackendEnv(srvURL string) func() {
	old := map[string]string{
		"MANAGEMENT_API_URL":   os.Getenv("MANAGEMENT_API_URL"),
		"KEYCLOAK_HOST":        os.Getenv("KEYCLOAK_HOST"),
		"KEYCLOAK_REALM":       os.Getenv("KEYCLOAK_REALM"),
		"RB_CDN_CLIENT_ID":     os.Getenv("RB_CDN_CLIENT_ID"),
		"RB_CDN_CLIENT_SECRET": os.Getenv("RB_CDN_CLIENT_SECRET"),
	}
	_ = os.Setenv("MANAGEMENT_API_URL", srvURL)
	_ = os.Setenv("KEYCLOAK_HOST", srvURL)
	_ = os.Setenv("KEYCLOAK_REALM", "rb")
	_ = os.Setenv("RB_CDN_CLIENT_ID", "rb-cdn-service-account")
	_ = os.Setenv("RB_CDN_CLIENT_SECRET", "secret")
	return func() {
		for k, v := range old {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	}
}

func TestSyncCapabilities_HappyPath(t *testing.T) {
	srv, syncCalls, lastBody := startStubBackend(t,
		http.StatusOK,
		`{"identity_id":"id-1","added":[],"updated":[],"deleted_ids":[]}`,
	)
	cleanup := setBackendEnv(srv.URL)
	defer cleanup()

	minioSvc := &fakeBucketLister{
		buckets: []minio.BucketInfo{{Name: "public-images"}, {Name: "videos"}},
	}
	rec := &recordingFatal{}

	SyncCapabilities(minioSvc, newTestLogger(), rec.Fn())

	if rec.called.Load() {
		t.Fatalf("fatal called unexpectedly: %s %+v", rec.format, rec.args)
	}
	if got := syncCalls.Load(); got != 1 {
		t.Errorf("syncCalls = %d, want 1", got)
	}
	body := lastBody.Load()
	if body == nil {
		t.Fatalf("no sync body captured")
	}
	// Pin: read + write base, plus 2 buckets * 2 verbs = 4 scoped → 6 total.
	if !strings.Contains(*body, `"capability":"read"`) || !strings.Contains(*body, `"capability":"write"`) {
		t.Errorf("body missing base capabilities: %s", *body)
	}
	if !strings.Contains(*body, `"scope":"bucket:public-images"`) || !strings.Contains(*body, `"scope":"bucket:videos"`) {
		t.Errorf("body missing per-bucket scopes: %s", *body)
	}
}

func TestSyncCapabilities_ListBucketsFailureIsFatal(t *testing.T) {
	srv, _, _ := startStubBackend(t, http.StatusOK, `{"identity_id":"x","added":[],"updated":[],"deleted_ids":[]}`)
	cleanup := setBackendEnv(srv.URL)
	defer cleanup()

	minioSvc := &fakeBucketLister{err: errors.ServiceError("minio is down")}
	rec := &recordingFatal{}

	SyncCapabilities(minioSvc, newTestLogger(), rec.Fn())

	if !rec.called.Load() {
		t.Fatalf("fatal not called")
	}
	if !strings.Contains(rec.format, "list buckets failed") {
		t.Errorf("fatal message = %q, want list-buckets context", rec.format)
	}
}

func TestSyncCapabilities_404IsFatalWithRegistrationHint(t *testing.T) {
	srv, _, _ := startStubBackend(t, http.StatusNotFound,
		`{"title":"Provider Identity not registered"}`)
	cleanup := setBackendEnv(srv.URL)
	defer cleanup()

	minioSvc := &fakeBucketLister{buckets: []minio.BucketInfo{{Name: "public-images"}}}
	rec := &recordingFatal{}

	SyncCapabilities(minioSvc, newTestLogger(), rec.Fn())

	if !rec.called.Load() {
		t.Fatalf("fatal not called")
	}
	if !strings.Contains(rec.format, "POST /v1/identities") {
		t.Errorf("fatal message should point operator at the fix, got %q", rec.format)
	}
}

func TestSyncCapabilities_500IsFatal(t *testing.T) {
	srv, _, _ := startStubBackend(t, http.StatusInternalServerError, `boom`)
	cleanup := setBackendEnv(srv.URL)
	defer cleanup()

	minioSvc := &fakeBucketLister{}
	rec := &recordingFatal{}

	SyncCapabilities(minioSvc, newTestLogger(), rec.Fn())

	if !rec.called.Load() {
		t.Fatalf("fatal not called on 500")
	}
}
