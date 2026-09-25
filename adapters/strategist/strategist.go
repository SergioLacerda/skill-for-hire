// Package strategist is the reference translation layer between the
// Strategist orchestrator's pilots (Cartografo, Jewelcrafter) and the
// Skills-for-Hire Knowledge API providers.
//
// It is deliberately provider-neutral: the pilot side declares which
// capability it needs, this package resolves a KnowledgeProvider
// implementation, and every response is rewrapped with Strategist-side
// policy (fallback, redaction) and observability. Skills never learn
// they're being consumed by Strategist.
//
// This scaffold ships two features:
//
//  1. A Router that picks a provider by capability rather than name.
//  2. A Piloted wrapper that adds policy-driven fallback and simple
//     telemetry hooks (event names only; the Strategist runtime
//     provides the actual sink).
//
// Wave 4c will add: read-through cache, per-mission budget accounting,
// and a live health-refresh loop.
package strategist

import (
	"context"
	"errors"
	"fmt"

	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
)

// PilotName is the Strategist-side label used only for observability.
type PilotName string

// Named pilots the Strategist runtime recognizes today.
const (
	PilotCartografo   PilotName = "cartografo"
	PilotJewelcrafter PilotName = "jewelcrafter"
)

// Router picks a KnowledgeProvider given a required capability. It is
// small on purpose — the Strategist runtime owns provider lifecycle
// (installation, upgrade, health), the router just remembers the
// mapping.
type Router struct {
	byCapability map[string]ka.KnowledgeProvider
}

// NewRouter returns an empty router. Register providers before use.
func NewRouter() *Router {
	return &Router{byCapability: make(map[string]ka.KnowledgeProvider)}
}

// Register wires a provider to every capability it advertises in Status.
// A second Register call for a capability replaces the previous one —
// this matches the Strategist runtime's "last-installed wins" rule.
func (r *Router) Register(ctx context.Context, p ka.KnowledgeProvider) error {
	if p == nil {
		return errors.New("strategist: cannot register nil provider")
	}
	s, err := p.Status(ctx)
	if err != nil {
		return fmt.Errorf("strategist: provider Status: %w", err)
	}
	if len(s.Capabilities) == 0 {
		return fmt.Errorf("strategist: provider %s@%s declares no capabilities", s.Provider, s.Version)
	}
	for _, cap := range s.Capabilities {
		r.byCapability[cap] = p
	}
	return nil
}

// Resolve returns the provider satisfying capability, or false when none
// is registered. The caller decides whether to fall back or fail.
func (r *Router) Resolve(capability string) (ka.KnowledgeProvider, bool) {
	p, ok := r.byCapability[capability]
	return p, ok
}

// TelemetrySink receives one event per provider call. Implementations
// are expected to be cheap and non-blocking — the Strategist runtime
// provides its own sink.
type TelemetrySink interface {
	// OnCall records that a provider handled (or refused) a request.
	// `err` is nil on success. Callers must not mutate `env` or `err`.
	OnCall(pilot PilotName, op string, env ka.Envelope, err error)
}

// NoopTelemetry drops every event. Useful for tests.
type NoopTelemetry struct{}

// OnCall satisfies TelemetrySink and does nothing.
func (NoopTelemetry) OnCall(PilotName, string, ka.Envelope, error) {}

// Policy encodes the Strategist-side rules a pilot enforces around
// provider calls. Empty policy = permissive (no fallback, telemetry
// noop, no source restriction).
type Policy struct {
	// AllowStaleReads, when false, converts a stale response into
	// ka.ErrStale so the caller can decide whether to refresh.
	AllowStaleReads bool
	// FailOnDegradedFallback, when true, refuses a response whose
	// envelope reports a non-none FallbackState. The pilot can then
	// pick a different provider or short-circuit the mission.
	FailOnDegradedFallback bool
}

// Piloted wraps a KnowledgeProvider with pilot-side policy + telemetry.
// It satisfies KnowledgeProvider itself so callers upstream see the
// same contract.
type Piloted struct {
	Pilot     PilotName
	Provider  ka.KnowledgeProvider
	Policy    Policy
	Telemetry TelemetrySink
}

// Wrap builds a Piloted with sensible defaults (noop telemetry).
func Wrap(pilot PilotName, p ka.KnowledgeProvider, policy Policy) *Piloted {
	return &Piloted{
		Pilot:     pilot,
		Provider:  p,
		Policy:    policy,
		Telemetry: NoopTelemetry{},
	}
}

// Prepare passes through and records telemetry.
func (w *Piloted) Prepare(ctx context.Context, req ka.PrepareRequest) (ka.PrepareResult, error) {
	res, err := w.Provider.Prepare(ctx, req)
	w.record("prepare", res.Envelope, err)
	return res, err
}

// Search enforces stale + fallback policy and records telemetry.
func (w *Piloted) Search(ctx context.Context, q ka.Query) (ka.SearchResult, error) {
	res, err := w.Provider.Search(ctx, q)
	w.record("search", res.Envelope, err)
	if err != nil {
		return res, err
	}
	if err := w.enforcePolicy(res.Envelope); err != nil {
		return res, err
	}
	return res, nil
}

// Refresh passes through and records telemetry.
func (w *Piloted) Refresh(ctx context.Context, scope ka.Scope) (ka.RefreshResult, error) {
	res, err := w.Provider.Refresh(ctx, scope)
	w.record("refresh", res.Envelope, err)
	return res, err
}

// Status passes through and records telemetry (envelope is synthesized
// from the status fields since Status doesn't carry an Envelope).
func (w *Piloted) Status(ctx context.Context) (ka.Status, error) {
	s, err := w.Provider.Status(ctx)
	env := ka.Envelope{
		Provider:         s.Provider,
		Version:          s.Version,
		SchemaVersion:    s.SchemaVersion,
		CapabilitiesUsed: s.Capabilities,
	}
	w.record("status", env, err)
	return s, err
}

// Explain passes through and records telemetry.
func (w *Piloted) Explain(ctx context.Context, itemID string) (ka.Explanation, error) {
	exp, err := w.Provider.Explain(ctx, itemID)
	w.record("explain", ka.Envelope{}, err)
	return exp, err
}

func (w *Piloted) record(op string, env ka.Envelope, err error) {
	sink := w.Telemetry
	if sink == nil {
		return
	}
	sink.OnCall(w.Pilot, op, env, err)
}

func (w *Piloted) enforcePolicy(env ka.Envelope) error {
	if !w.Policy.AllowStaleReads && env.Freshness == ka.FreshnessStale {
		return ka.ErrStale
	}
	if w.Policy.FailOnDegradedFallback && env.FallbackState != "" && env.FallbackState != ka.FallbackNone {
		return ka.ErrFallbackTriggered
	}
	return nil
}

// Compile-time proof that Piloted still satisfies the KA contract.
var _ ka.KnowledgeProvider = (*Piloted)(nil)
