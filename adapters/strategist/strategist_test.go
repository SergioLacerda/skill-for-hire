package strategist_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SergioLacerda/skill-for-hire/adapters/strategist"
	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
)

type fakeProvider struct {
	prepared     bool
	freshness    ka.Freshness
	fallback     ka.FallbackState
	capabilities []string
}

func (f *fakeProvider) Prepare(context.Context, ka.PrepareRequest) (ka.PrepareResult, error) {
	f.prepared = true
	return ka.PrepareResult{Envelope: f.envelope()}, nil
}

func (f *fakeProvider) Search(context.Context, ka.Query) (ka.SearchResult, error) {
	if !f.prepared {
		return ka.SearchResult{}, ka.ErrNotPrepared
	}
	return ka.SearchResult{Envelope: f.envelope()}, nil
}

func (f *fakeProvider) Refresh(context.Context, ka.Scope) (ka.RefreshResult, error) {
	return ka.RefreshResult{Envelope: f.envelope()}, nil
}

func (f *fakeProvider) Status(context.Context) (ka.Status, error) {
	return ka.Status{
		Provider:      "fake",
		Version:       "1.0.0",
		SchemaVersion: ka.SchemaVersion,
		Healthy:       true,
		Capabilities:  f.capabilities,
	}, nil
}

func (f *fakeProvider) Explain(_ context.Context, itemID string) (ka.Explanation, error) {
	return ka.Explanation{ItemID: itemID, Decision: ka.DecisionSelected}, nil
}

func (f *fakeProvider) envelope() ka.Envelope {
	return ka.Envelope{
		Provider:         "fake",
		Version:          "1.0.0",
		SchemaVersion:    ka.SchemaVersion,
		CapabilitiesUsed: f.capabilities,
		Freshness:        f.freshness,
		FallbackState:    f.fallback,
	}
}

func TestRouterResolvesByCapability(t *testing.T) {
	r := strategist.NewRouter()
	p := &fakeProvider{capabilities: []string{"knowledge.search", "knowledge.mine"}}
	if err := r.Register(context.Background(), p); err != nil {
		t.Fatalf("register: %v", err)
	}
	for _, cap := range []string{"knowledge.search", "knowledge.mine"} {
		if got, ok := r.Resolve(cap); !ok || got == nil {
			t.Errorf("Resolve(%q) missed the provider", cap)
		}
	}
	if _, ok := r.Resolve("architecture.map"); ok {
		t.Error("Resolve returned a provider for an unregistered capability")
	}
}

func TestRegisterRejectsNoCapabilityProvider(t *testing.T) {
	r := strategist.NewRouter()
	if err := r.Register(context.Background(), &fakeProvider{capabilities: nil}); err == nil {
		t.Fatal("expected error registering a provider with no declared capabilities")
	}
}

func TestPolicyBlocksStaleWhenDisallowed(t *testing.T) {
	p := &fakeProvider{
		capabilities: []string{"knowledge.search"},
		freshness:    ka.FreshnessStale,
	}
	_, _ = p.Prepare(context.Background(), ka.PrepareRequest{})
	piloted := strategist.Wrap(strategist.PilotJewelcrafter, p, strategist.Policy{AllowStaleReads: false})

	_, err := piloted.Search(context.Background(), ka.Query{})
	if !errors.Is(err, ka.ErrStale) {
		t.Fatalf("expected ErrStale, got %v", err)
	}
}

func TestPolicyAllowsStaleWhenPermitted(t *testing.T) {
	p := &fakeProvider{
		capabilities: []string{"knowledge.search"},
		freshness:    ka.FreshnessStale,
	}
	_, _ = p.Prepare(context.Background(), ka.PrepareRequest{})
	piloted := strategist.Wrap(strategist.PilotJewelcrafter, p, strategist.Policy{AllowStaleReads: true})

	if _, err := piloted.Search(context.Background(), ka.Query{}); err != nil {
		t.Fatalf("expected stale-permitted search to pass, got %v", err)
	}
}

func TestPolicyFailsOnDegradedFallback(t *testing.T) {
	p := &fakeProvider{
		capabilities: []string{"knowledge.search"},
		fallback:     ka.FallbackDegraded,
	}
	_, _ = p.Prepare(context.Background(), ka.PrepareRequest{})
	piloted := strategist.Wrap(strategist.PilotJewelcrafter, p, strategist.Policy{
		AllowStaleReads:        true,
		FailOnDegradedFallback: true,
	})

	_, err := piloted.Search(context.Background(), ka.Query{})
	if !errors.Is(err, ka.ErrFallbackTriggered) {
		t.Fatalf("expected ErrFallbackTriggered, got %v", err)
	}
}

type recordingTelemetry struct {
	events []string
}

func (r *recordingTelemetry) OnCall(pilot strategist.PilotName, op string, _ ka.Envelope, _ error) {
	r.events = append(r.events, string(pilot)+":"+op)
}

func TestTelemetryReceivesEveryCall(t *testing.T) {
	p := &fakeProvider{
		capabilities: []string{"knowledge.search"},
		freshness:    ka.FreshnessFresh,
	}
	sink := &recordingTelemetry{}
	piloted := strategist.Wrap(strategist.PilotJewelcrafter, p, strategist.Policy{AllowStaleReads: true})
	piloted.Telemetry = sink

	if _, err := piloted.Prepare(context.Background(), ka.PrepareRequest{Root: "."}); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if _, err := piloted.Search(context.Background(), ka.Query{}); err != nil {
		t.Fatalf("search: %v", err)
	}
	if _, err := piloted.Refresh(context.Background(), ka.Scope{}); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if _, err := piloted.Status(context.Background()); err != nil {
		t.Fatalf("status: %v", err)
	}
	if _, err := piloted.Explain(context.Background(), "x"); err != nil {
		t.Fatalf("explain: %v", err)
	}
	want := []string{
		"jewelcrafter:prepare",
		"jewelcrafter:search",
		"jewelcrafter:refresh",
		"jewelcrafter:status",
		"jewelcrafter:explain",
	}
	if len(sink.events) != len(want) {
		t.Fatalf("event count: got %d, want %d — events: %v", len(sink.events), len(want), sink.events)
	}
	for i, w := range want {
		if sink.events[i] != w {
			t.Errorf("event %d = %q, want %q", i, sink.events[i], w)
		}
	}
}
