package runtime

import (
	"testing"
)

func TestDiffSignatures_AddedRemovedUpdated(t *testing.T) {
	old := map[string]string{
		"a": "v1",
		"b": "v1",
		"c": "v1",
	}
	newSig := map[string]string{
		"a": "v1", // unchanged
		"b": "v2", // updated
		// c removed
		"d": "v1", // added
	}
	added, updated, removed := diffSignatures(old, newSig)
	if added != 1 {
		t.Errorf("added = %d, want 1", added)
	}
	if updated != 1 {
		t.Errorf("updated = %d, want 1", updated)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}
}

func TestDiffSignatures_EmptyBothZero(t *testing.T) {
	added, updated, removed := diffSignatures(map[string]string{}, map[string]string{})
	if added != 0 || updated != 0 || removed != 0 {
		t.Fatalf("expected all zero, got a=%d u=%d r=%d", added, updated, removed)
	}
}

func TestDiffSignatures_AllAddedFromEmpty(t *testing.T) {
	added, updated, removed := diffSignatures(map[string]string{}, map[string]string{"x": "v", "y": "v"})
	if added != 2 || updated != 0 || removed != 0 {
		t.Fatalf("added=%d updated=%d removed=%d; want 2/0/0", added, updated, removed)
	}
}
