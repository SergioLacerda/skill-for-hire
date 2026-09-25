package runtime

import (
	"strings"

	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
)

// trustToFloat maps the domain's textual trust levels to the [0, 1] range
// KA envelopes use. Unknown values fall to 0 (least trusted) so a
// malformed manifest can never masquerade as high-trust.
func trustToFloat(trust string) float64 {
	switch strings.ToLower(strings.TrimSpace(trust)) {
	case "high":
		return 0.9
	case "medium":
		return 0.6
	case "low":
		return 0.3
	default:
		return 0.0
	}
}

// jewelMatch computes a relevance score for a jewel against a query and
// returns whether the jewel matched at all.
//
// Ranking policy (deliberately explicit — see docs/architecture, §5.5
// "explainability of ranking"):
//
//   - Statement match  : +3
//   - Kind match       : +1
//   - Applicability.Scope match:       +2 (semantic, more valuable than kind)
//   - Applicability.AppliesWhen match: +2
//   - Applicability.AvoidWhen match:   -4 (strong penalty — the jewel says
//     don't apply here)
//   - Curatorial JewelScore.Value bonus: added to relevance so a
//     well-curated jewel wins ties over a poorly-scored one on the same
//     match count.
//
// Empty needles = match all (score 1 baseline). AvoidWhen still penalises.
func jewelMatch(j Jewel, needles []string) (score float64, matched bool) {
	if len(needles) == 0 {
		matched = true
		score = 1.0
	}

	for _, n := range needles {
		hit := false
		if containsFold(j.Statement, n) {
			score += 3
			hit = true
		}
		if containsFold(j.Kind, n) {
			score++
			hit = true
		}
		if anyContainsFold(j.Applicability.Scope, n) {
			score += 2
			hit = true
		}
		if anyContainsFold(j.Applicability.AppliesWhen, n) {
			score += 2
			hit = true
		}
		if anyContainsFold(j.Applicability.AvoidWhen, n) {
			score -= 4
		}
		if hit {
			matched = true
		}
	}

	// Curatorial score contributes as a small tie-breaker rather than
	// dominating: a Value of 100 adds 1.0 relevance.
	score += float64(j.Score.Value) / 100.0

	return score, matched
}

// potionMatch is the potion analogue. Potions have less structured
// metadata than jewels, so the weights are lighter.
func potionMatch(p Potion, needles []string) (score float64, matched bool) {
	if len(needles) == 0 {
		matched = true
		score = 1.0
	}
	for _, n := range needles {
		hit := false
		if containsFold(p.WhenToUse, n) {
			score += 3
			hit = true
		}
		if containsFold(p.WhenToAvoid, n) {
			score -= 4
		}
		if containsFold(p.RunbookRef, n) {
			score++
			hit = true
		}
		if hit {
			matched = true
		}
	}
	return score, matched
}

// rankItemsInPlace sorts by Score descending, tie-broken by ID for
// determinism (KA envelopes are compared byte-for-byte by consumers'
// lockfile flow).
func rankItemsInPlace(items []ka.Item) {
	// Simple in-place insertion sort — items is small (bounded by
	// query.TokenBudget or repo size), and this keeps the ranking
	// deterministic without dragging a sort.Slice closure that risks
	// panicking on NaN scores.
	for i := 1; i < len(items); i++ {
		j := i
		for j > 0 && lessItem(items[j], items[j-1]) {
			items[j], items[j-1] = items[j-1], items[j]
			j--
		}
	}
}

func lessItem(a, b ka.Item) bool {
	if a.Score != b.Score {
		return a.Score > b.Score // descending
	}
	if a.Trust != b.Trust {
		return a.Trust > b.Trust
	}
	return a.ID < b.ID
}

func containsFold(haystack, needle string) bool {
	if needle == "" || haystack == "" {
		return false
	}
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

func anyContainsFold(hay []string, needle string) bool {
	for _, h := range hay {
		if containsFold(h, needle) {
			return true
		}
	}
	return false
}
