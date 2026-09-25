package knowledgeapi

import "errors"

// Normalized errors surfaced across every provider. Consumers switch on
// these sentinels to route (retry, fall back, block); provider-specific
// messages wrap them with %w.
var (
	// ErrNotPrepared means Prepare has not yet run for this workspace.
	ErrNotPrepared = errors.New("knowledgeapi: provider not prepared; call Prepare first")

	// ErrIncompatibleVersion means the caller's schema version and the
	// provider's do not agree. Do not partially parse — reject.
	ErrIncompatibleVersion = errors.New("knowledgeapi: incompatible schema version")

	// ErrPermissionDenied means the provider refused the requested
	// scope (root outside allowlist, secrets access, subprocess).
	ErrPermissionDenied = errors.New("knowledgeapi: permission denied")

	// ErrSourceMissing means an expected source under the current
	// snapshot is absent. Consumers may Refresh and retry.
	ErrSourceMissing = errors.New("knowledgeapi: source missing")

	// ErrStale means the answer would be served from a stale index and
	// the caller did not authorize stale reads. Consumers may Refresh.
	ErrStale = errors.New("knowledgeapi: index is stale and stale reads disabled")

	// ErrFallbackTriggered is returned WITH a valid envelope when the
	// provider had to fall back. The envelope's FallbackState names
	// which flavor; callers decide whether the degraded answer is
	// acceptable.
	ErrFallbackTriggered = errors.New("knowledgeapi: fallback triggered")

	// ErrUnavailable means the provider is not currently reachable or
	// its Status reports unhealthy. Not a routing decision — the
	// provider itself is down.
	ErrUnavailable = errors.New("knowledgeapi: provider unavailable")
)
