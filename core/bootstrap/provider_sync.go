package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RodolfoBonis/rb-cdn/core/config"
	apperrors "github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb_auth_client/provider"
	"github.com/minio/minio-go"
)

// BucketLister is the slice of services.IMinioService that this
// boot phase actually depends on. Narrowed so tests don't have to
// stub out the full upload/download surface to exercise the sync.
type BucketLister interface {
	ListBuckets() ([]minio.BucketInfo, *apperrors.AppError)
}

// SyncTimeout caps the boot-time round-trip to the management API.
// k8s gives the pod plenty of time to start; we'd rather fail fast
// than hang the boot if mgmt-api is slow or unreachable.
const SyncTimeout = 30 * time.Second

// FatalFunc is the abort path called on any sync failure. main.go
// wires this to log.Fatalf; tests inject a recording stub so they
// can assert which path triggered the abort without killing the
// test process.
type FatalFunc func(format string, args ...any)

// SyncCapabilities reconciles rb-cdn's declared capabilities with
// the management API catalog at boot, fail-closed: any failure
// (MinIO unreachable, mgmt-api down, provider unregistered)
// terminates the process via fatal.
//
// Capability shape:
//   - "read" / "write" — service-level base capabilities. These
//     match what RequireServicePermission("rb-cdn", "read"|"write")
//     enforces today.
//   - "read" / "write" with scope = "bucket:<name>" — declared per
//     bucket the configured MinIO credentials can see at boot. New
//     buckets created after boot won't appear in the catalog until
//     the next deploy; acceptable v1 trade-off given rb-cdn's
//     short deploy cycle. The bucket-level permission check at
//     request time is unchanged.
//
// Pre-condition: the rb-cdn service Identity (client_id matches
// RB_CDN_CLIENT_ID) must already be registered in rb_management_api
// via POST /v1/identities. The Sync endpoint returns 404 otherwise;
// this function maps that to a clear fatal message pointing at the
// fix.
func SyncCapabilities(minioSvc BucketLister, log *logger.CustomLogger, fatal FatalFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), SyncTimeout)
	defer cancel()

	buckets, appErr := minioSvc.ListBuckets()
	if appErr != nil {
		fatal("provider sync: list buckets failed: %s", appErr.Message)
		return
	}

	p, err := provider.New(provider.Config{
		Name:         "rb-cdn",
		MgmtAPIURL:   config.EnvManagementAPIURL(),
		KeycloakURL:  config.EnvKeycloakHost(),
		Realm:        config.EnvKeycloakRealm(),
		ClientID:     config.EnvRBCDNClientID(),
		ClientSecret: config.EnvRBCDNClientSecret(),
		Logger:       NewRBAuthLogger(log),
		HTTPTimeout:  10 * time.Second,
	})
	if err != nil {
		fatal("provider sync: New: %v", err)
		return
	}

	// Service-level base. Mirrors the literal middleware checks at
	// route registration (`RequireServicePermission("rb-cdn", "read"|"write")`).
	p.Capability("read")
	p.Capability("write")

	// Per-bucket scopes. Empty bucket list is legitimate (fresh
	// MinIO with no data yet), so we don't fail on an empty list —
	// only if ListBuckets itself errored above.
	for _, b := range buckets {
		bucketScope := "bucket:" + b.Name
		p.Capability("read").Scope(bucketScope)
		p.Capability("write").Scope(bucketScope)
	}

	res, err := p.Sync(ctx)
	if err != nil {
		// Translate the unregistered-provider sentinel into a
		// concrete operator instruction. Anything else is logged as
		// a generic sync failure.
		if errors.Is(err, provider.ErrUnregisteredProvider) {
			fatal("provider sync: identity %q is not registered in rb_management_api — "+
				"register it via POST /v1/identities before deploying. underlying: %v",
				config.EnvRBCDNClientID(), err)
			return
		}
		fatal("provider sync: %v", err)
		return
	}

	log.Info("provider sync: capabilities reconciled", map[string]interface{}{
		"identity_id":  res.IdentityID,
		"buckets_seen": len(buckets),
		"added":        len(res.Added),
		"updated":      len(res.Updated),
		"deleted":      len(res.DeletedIDs),
	})
}

// DefaultFatal is the production abort path. Wraps log.Fatalf
// equivalents in fmt.Errorf semantics so the message format is
// identical to other rb-cdn boot failures.
func DefaultFatal(log *logger.CustomLogger) FatalFunc {
	return func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		log.Error(msg, map[string]interface{}{"phase": "boot"})
		// Direct panic so the surrounding init() / main() flow
		// terminates the process — k8s will restart the pod and the
		// next attempt either succeeds (transient mgmt-api hiccup)
		// or surfaces the same loud message in CrashLoopBackOff
		// logs.
		panic("rb-cdn boot aborted: " + msg)
	}
}
