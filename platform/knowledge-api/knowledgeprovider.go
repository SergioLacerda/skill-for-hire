// Package knowledgeapi is the provider-neutral contract every knowledge
// skill implements. Consumers (Strategist adapter, standalone CLIs, other
// agents) route by capability rather than by provider name, so no
// implementation lives here — only interface, request/response shapes,
// enums and normalized errors.
//
// Deliberately dependency-free (stdlib only): the contract must not drag
// third-party types into every consumer.
package knowledgeapi

import "context"

// SchemaVersion is the current major of the Knowledge API contract.
// Consumers compare against this to reject incompatible providers
// loudly rather than reading garbage silently.
const SchemaVersion = 1

// KnowledgeProvider is the surface skills like atlas and treasure-chest
// must expose. Implementations belong to individual skills; this package
// only defines the contract.
//
// Every method returns an Envelope so consumers get provider identity,
// capabilities used, source provenance, freshness, trust, cost, and
// fallback state on every call — never just "here is a result".
type KnowledgeProvider interface {
	// Prepare seeds the provider from authorized sources. It is a heavy
	// operation and MUST be explicit — consumers never trigger it as a
	// side effect of Search/Refresh.
	Prepare(ctx context.Context, req PrepareRequest) (PrepareResult, error)

	// Search returns candidates ranked by applicability. Items carry
	// evidence and an Envelope; consumers select what to load fully.
	Search(ctx context.Context, q Query) (SearchResult, error)

	// Refresh reindexes only what changed inside the given scope. Full
	// rebuild belongs to Prepare; Refresh MUST be incremental.
	Refresh(ctx context.Context, scope Scope) (RefreshResult, error)

	// Status reports health, current schema, capabilities actually
	// available, and last-prepare timestamp. Consumers check it before
	// routing; "registered" is not "healthy".
	Status(ctx context.Context) (Status, error)

	// Explain answers "why was this item selected / discarded /
	// considered incompatible?" — the AI-first requirement that no
	// ranking is opaque.
	Explain(ctx context.Context, itemID string) (Explanation, error)
}

// PrepareRequest is the input to Prepare.
type PrepareRequest struct {
	// Root is the workspace root the provider is authorized to read.
	Root string
	// Sources restricts what to ingest (globs, paths). Empty = all
	// discoverable sources.
	Sources []string
	// Policy carries optional trust/scope/TTL knobs.
	Policy Policy
}

// PrepareResult reports what Prepare did.
type PrepareResult struct {
	Envelope     Envelope
	SourcesRead  int
	ItemsIndexed int
}

// Query is the input to Search.
type Query struct {
	Intent      string
	Concepts    []string
	Symbols     []string
	Paths       []string
	Components  []string
	TokenBudget int
}

// SearchResult is the output of Search.
type SearchResult struct {
	Envelope Envelope
	Items    []Item
}

// Item is one candidate returned by Search.
type Item struct {
	ID            string
	Kind          string
	Score         float64
	Trust         float64
	Freshness     Freshness
	Applicability string
	Reason        string
	Source        Source
}

// Scope is the input to Refresh — the region to reindex.
type Scope struct {
	Root  string
	Since string // git sha; empty means "since last snapshot"
	Paths []string
}

// RefreshResult is the output of Refresh.
type RefreshResult struct {
	Envelope Envelope
	Added    int
	Updated  int
	Removed  int
}

// Status is the output of Status.
type Status struct {
	Provider      string
	Version       string
	SchemaVersion int
	Healthy       bool
	Capabilities  []string
	LastPrepareAt string // RFC3339 or empty
	Detail        string // optional human-readable note
}

// Explanation is the output of Explain.
type Explanation struct {
	ItemID   string
	Decision Decision
	Reason   string
	Evidence []Source
}

// Policy carries optional knobs shared by all methods.
type Policy struct {
	TrustFloor    float64
	ScopeFilter   ScopeFilter
	FreshnessTTL  string // ISO 8601 duration; empty means default
	AllowFallback bool
}
