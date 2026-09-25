package knowledgeapi_test

import (
	"context"
	"errors"
	"testing"

	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
)

// fakeProvider is a compile-time proof that the interface is
// implementable in ~40 lines without pulling in any domain. Every
// method returns an Envelope so the "responses carry provenance" rule
// is enforced by the type system, not just documentation.
type fakeProvider struct {
	prepared bool
}

func (f *fakeProvider) Prepare(_ context.Context, _ ka.PrepareRequest) (ka.PrepareResult, error) {
	f.prepared = true
	return ka.PrepareResult{
		Envelope:     f.envelope(nil),
		SourcesRead:  0,
		ItemsIndexed: 0,
	}, nil
}

func (f *fakeProvider) Search(_ context.Context, _ ka.Query) (ka.SearchResult, error) {
	if !f.prepared {
		return ka.SearchResult{}, ka.ErrNotPrepared
	}
	return ka.SearchResult{Envelope: f.envelope(nil)}, nil
}

func (f *fakeProvider) Refresh(_ context.Context, _ ka.Scope) (ka.RefreshResult, error) {
	return ka.RefreshResult{Envelope: f.envelope(nil)}, nil
}

func (f *fakeProvider) Status(_ context.Context) (ka.Status, error) {
	return ka.Status{
		Provider:      "fake",
		Version:       "0.0.0",
		SchemaVersion: ka.SchemaVersion,
		Healthy:       true,
		Capabilities:  []string{"test.echo"},
	}, nil
}

func (f *fakeProvider) Explain(_ context.Context, itemID string) (ka.Explanation, error) {
	return ka.Explanation{
		ItemID:   itemID,
		Decision: ka.DecisionSelected,
		Reason:   "fake always selects",
	}, nil
}

func (f *fakeProvider) envelope(limits []string) ka.Envelope {
	return ka.Envelope{
		Provider:         "fake",
		Version:          "0.0.0",
		SchemaVersion:    ka.SchemaVersion,
		CapabilitiesUsed: []string{"test.echo"},
		Freshness:        ka.FreshnessFresh,
		Trust:            1.0,
		FallbackState:    ka.FallbackNone,
		Limitations:      limits,
	}
}

// TestInterfaceIsImplementable is the compile-time guard: if
// KnowledgeProvider ever changes shape, this file stops compiling
// before the failure ships as a broken contract to atlas or
// treasure-chest. The parameter is intentionally unused — the check
// runs at compile time via the blank var declaration below.
func TestInterfaceIsImplementable(_ *testing.T) {
	var _ ka.KnowledgeProvider = (*fakeProvider)(nil)
}

func TestSearchWithoutPrepareReturnsSentinel(t *testing.T) {
	p := &fakeProvider{}
	_, err := p.Search(context.Background(), ka.Query{})
	if !errors.Is(err, ka.ErrNotPrepared) {
		t.Fatalf("expected ErrNotPrepared, got %v", err)
	}
}

func TestPrepareEnablesSearch(t *testing.T) {
	p := &fakeProvider{}
	if _, err := p.Prepare(context.Background(), ka.PrepareRequest{Root: "."}); err != nil {
		t.Fatalf("prepare failed: %v", err)
	}
	res, err := p.Search(context.Background(), ka.Query{Intent: "anything"})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if res.Envelope.Provider == "" {
		t.Fatal("response envelope missing provider identity")
	}
	if res.Envelope.SchemaVersion != ka.SchemaVersion {
		t.Fatalf("schema version drift: got %d, want %d", res.Envelope.SchemaVersion, ka.SchemaVersion)
	}
	if res.Envelope.FallbackState != ka.FallbackNone {
		t.Fatalf("unexpected fallback state: %s", res.Envelope.FallbackState)
	}
}

func TestStatusReportsCapabilities(t *testing.T) {
	p := &fakeProvider{}
	s, err := p.Status(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !s.Healthy {
		t.Fatal("fake should be healthy")
	}
	if len(s.Capabilities) == 0 {
		t.Fatal("status must list capabilities; \"registered\" is not \"healthy\"")
	}
}
