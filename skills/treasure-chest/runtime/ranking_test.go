package runtime

import (
	"testing"

	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
)

func TestJewelMatch_StatementBeatsKind(t *testing.T) {
	// Two jewels, both match on kind, only one matches on statement.
	// The statement match must rank higher (weight 3 vs 1).
	j1 := Jewel{ID: "stmt", Kind: "runbook", Statement: "handle dependency upgrade carefully"}
	j2 := Jewel{ID: "kind", Kind: "runbook", Statement: "some other thing"}

	s1, m1 := jewelMatch(j1, []string{"runbook", "dependency"})
	s2, m2 := jewelMatch(j2, []string{"runbook"})

	if !m1 || !m2 {
		t.Fatalf("both should match; m1=%v m2=%v", m1, m2)
	}
	if s1 <= s2 {
		t.Fatalf("expected statement match to beat kind-only: %f !> %f", s1, s2)
	}
}

func TestJewelMatch_AvoidWhenPenaltyOverridesMildBoost(t *testing.T) {
	// A jewel that says "avoid_when: production" and matches only on Kind
	// should end up worse than a no-match jewel.
	target := Jewel{
		Kind: "runbook",
		Applicability: JewelApplicability{
			AvoidWhen: []string{"production"},
		},
	}
	score, matched := jewelMatch(target, []string{"runbook", "production"})
	if !matched {
		t.Fatal("jewel should still register as matched (kind hit)")
	}
	// +1 (kind) -4 (avoid_when) = -3
	if score >= 0 {
		t.Fatalf("expected penalty to drive score negative, got %f", score)
	}
}

func TestJewelMatch_CuratorialScoreTieBreaker(t *testing.T) {
	// Same match count, different curatorial scores: higher Value wins.
	low := Jewel{ID: "low", Statement: "abc", Score: JewelScore{Value: 10}}
	high := Jewel{ID: "high", Statement: "abc", Score: JewelScore{Value: 90}}

	sLow, _ := jewelMatch(low, []string{"abc"})
	sHigh, _ := jewelMatch(high, []string{"abc"})
	if sHigh <= sLow {
		t.Fatalf("higher curatorial value must produce higher relevance: %f !> %f", sHigh, sLow)
	}
}

func TestRankItemsInPlace_DeterministicOrdering(t *testing.T) {
	items := []ka.Item{
		{ID: "b", Score: 1.5, Trust: 0.5},
		{ID: "a", Score: 1.5, Trust: 0.5},
		{ID: "c", Score: 3.0, Trust: 0.9},
		{ID: "d", Score: 2.0, Trust: 0.3},
	}
	rankItemsInPlace(items)
	want := []string{"c", "d", "a", "b"} // c highest; a<b for tie
	for i, w := range want {
		if items[i].ID != w {
			t.Fatalf("position %d: got %q want %q; full order: %v", i, items[i].ID, w, itemIDs(items))
		}
	}
}

func TestTrustToFloat_RangeAndFallback(t *testing.T) {
	cases := map[string]float64{
		"high":    0.9,
		"HIGH ":   0.9,
		"medium":  0.6,
		"low":     0.3,
		"":        0.0,
		"garbage": 0.0,
	}
	for in, want := range cases {
		if got := trustToFloat(in); got != want {
			t.Errorf("trustToFloat(%q) = %f, want %f", in, got, want)
		}
	}
}

func itemIDs(items []ka.Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.ID
	}
	return out
}
