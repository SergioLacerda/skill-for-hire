package knowledgeapi

// Envelope is the metadata every response carries. Consumers use it to
// decide trust, freshness, cost budget and fallback disclosure without
// re-inspecting the provider — the response speaks for itself.
type Envelope struct {
	// Provider identifies the skill and version that answered.
	Provider string
	Version  string

	// SchemaVersion is what the response conforms to. Consumers that
	// only understand older schemas MUST reject a higher value rather
	// than partially parse it.
	SchemaVersion int

	// CapabilitiesUsed is the intersection of what was requested and
	// what the provider actually served — never a superset.
	CapabilitiesUsed []string

	// Sources are the origins of the evidence in the response, each
	// with a digest so consumers can verify what they read.
	Sources []Source

	// Freshness reports whether the index behind this response is
	// current relative to its authoritative source.
	Freshness Freshness

	// Trust is the aggregate trust score of the response, in [0, 1].
	Trust float64

	// Compatibility carries the version tuple (see the ADR on ORKA /
	// project adapter / runtime): "package/x.y.z; adapter/x.y.z;
	// runtime/x.y.z". Empty when the caller did not negotiate one.
	Compatibility string

	// Cost is what this response cost — estimated when the call
	// finishes optimistically, actual after execution.
	Cost Cost

	// Limitations names anything the response is NOT (partial results,
	// dropped candidates, missing sources). Silence is a lie.
	Limitations []string

	// FallbackState describes what surfaced this answer: the live
	// provider, an embedded fallback, or a degraded mode.
	FallbackState FallbackState
}

// Source is the origin of one piece of evidence.
type Source struct {
	Path   string
	Digest string // sha256:...
	Commit string // optional git sha
}

// Cost is a lightweight, unit-agnostic pair. Concrete providers pick
// what the numbers mean (tokens, milliseconds, IO ops) and document it
// in their SKILL.md.
type Cost struct {
	Estimated int64
	Actual    int64
}

// Freshness reports index health relative to its source.
type Freshness string

// Freshness values. Consumers that see anything outside this set MUST
// reject the response rather than partially trust it.
const (
	FreshnessFresh          Freshness = "fresh"
	FreshnessPartiallyStale Freshness = "partially_stale"
	FreshnessStale          Freshness = "stale"
	FreshnessUnknown        Freshness = "unknown"
)

// FallbackState is how the answer was surfaced.
type FallbackState string

// FallbackState values.
const (
	FallbackNone     FallbackState = "none"
	FallbackEmbedded FallbackState = "embedded"
	FallbackDegraded FallbackState = "degraded"
)

// Decision is what Explain returns for one item.
type Decision string

// Decision values.
const (
	DecisionSelected     Decision = "selected"
	DecisionDiscarded    Decision = "discarded"
	DecisionIncompatible Decision = "incompatible"
)

// ScopeFilter narrows Prepare/Refresh/Search inputs.
type ScopeFilter string

// ScopeFilter values.
const (
	ScopeGlobal    ScopeFilter = "global"
	ScopeProduct   ScopeFilter = "product"
	ScopeWorkspace ScopeFilter = "workspace"
)
