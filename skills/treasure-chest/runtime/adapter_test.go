package runtime_test

import (
	"context"
	"errors"
	"testing"

	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
	"github.com/SergioLacerda/skill-for-hire/skills/treasure-chest/runtime"
)

// TestAdapterSatisfiesKnowledgeProvider is the compile-time guard.
func TestAdapterSatisfiesKnowledgeProvider(_ *testing.T) {
	var _ ka.KnowledgeProvider = (*runtime.Adapter)(nil)
}

func TestAdapterPrepareRequiresRoot(t *testing.T) {
	a := runtime.NewAdapter("0.2.0")
	if _, err := a.Prepare(context.Background(), ka.PrepareRequest{}); err == nil {
		t.Fatal("expected error when Root is empty")
	}
}

func TestAdapterSearchWithoutPrepareReturnsSentinel(t *testing.T) {
	a := runtime.NewAdapter("0.2.0")
	_, err := a.Search(context.Background(), ka.Query{Intent: "anything"})
	if !errors.Is(err, ka.ErrNotPrepared) {
		t.Fatalf("expected ErrNotPrepared, got %v", err)
	}
}

func TestAdapterStatusBeforePrepareIsUnhealthy(t *testing.T) {
	a := runtime.NewAdapter("0.2.0")
	s, err := a.Status(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if s.Provider != "treasure-chest" {
		t.Fatalf("provider identity: got %q", s.Provider)
	}
	if s.Version != "0.2.0" {
		t.Fatalf("version: got %q", s.Version)
	}
	if s.Healthy {
		t.Fatal("Healthy should be false before Prepare")
	}
	if len(s.Capabilities) == 0 {
		t.Fatal("Status must always list capabilities")
	}
	if s.SchemaVersion != ka.SchemaVersion {
		t.Fatalf("SchemaVersion drift: got %d, want %d", s.SchemaVersion, ka.SchemaVersion)
	}
}

func TestAdapterExplainUnknownIDReturnsDiscarded(t *testing.T) {
	a := runtime.NewAdapter("0.2.0")
	// Prepare against an empty temp dir so the index is real but empty.
	tmp := t.TempDir()
	if _, err := a.Prepare(context.Background(), ka.PrepareRequest{Root: tmp}); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	exp, err := a.Explain(context.Background(), "nope")
	if err != nil {
		t.Fatalf("explain: %v", err)
	}
	if exp.Decision != ka.DecisionDiscarded {
		t.Fatalf("expected Discarded for unknown id, got %s", exp.Decision)
	}
}
