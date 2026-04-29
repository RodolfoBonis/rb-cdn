// Package bootstrap holds initialisation routines that run during
// the rb-cdn boot sequence — things that must complete (or fail
// loudly) before the HTTP listener accepts traffic.
package bootstrap

import (
	"fmt"

	"github.com/RodolfoBonis/rb-cdn/core/logger"
)

// rbauthLoggerAdapter bridges rb-cdn's zap-backed CustomLogger to
// the slog-style Logger interface that rb_auth_client expects
// (Debug/Info/Warn/Error with variadic key-value args).
//
// The CustomLogger underneath wants `(message string, jsonData
// ...map[string]interface{})`, so we collapse the alternating
// key/value args into a single map per call. Non-string keys and
// odd-length args are tolerated (loud-but-safe) — the SDK's args
// always pair correctly today, but we don't want a runtime panic
// here regressing the whole boot.
type rbauthLoggerAdapter struct{ inner *logger.CustomLogger }

// NewRBAuthLogger wraps the global rb-cdn logger so the SDK can
// emit its own structured events through the same pipeline.
func NewRBAuthLogger(inner *logger.CustomLogger) *rbauthLoggerAdapter {
	return &rbauthLoggerAdapter{inner: inner}
}

// Debug — CustomLogger has no separate Debug level, so we route
// through Info. Acceptable trade-off: the SDK uses Debug only for
// noisy lifecycle events that aren't critical to surface separately.
func (a *rbauthLoggerAdapter) Debug(msg string, args ...any) {
	a.inner.Info(msg, kvToMap(args))
}

func (a *rbauthLoggerAdapter) Info(msg string, args ...any) {
	a.inner.Info(msg, kvToMap(args))
}

func (a *rbauthLoggerAdapter) Warn(msg string, args ...any) {
	a.inner.Warning(msg, kvToMap(args))
}

func (a *rbauthLoggerAdapter) Error(msg string, args ...any) {
	a.inner.Error(msg, kvToMap(args))
}

// kvToMap converts slog-style alternating (key, value, key, value,
// ...) args into a single map suitable for the CustomLogger. Keys
// are coerced to strings via fmt.Sprint so the conversion never
// drops a value. Odd-length input has its trailing arg recorded
// under the conventional "!BADKEY" sentinel (matches slog).
func kvToMap(args []any) map[string]interface{} {
	if len(args) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(args)/2)
	i := 0
	for ; i+1 < len(args); i += 2 {
		key := fmt.Sprint(args[i])
		out[key] = args[i+1]
	}
	if i < len(args) {
		out["!BADKEY"] = args[i]
	}
	return out
}
