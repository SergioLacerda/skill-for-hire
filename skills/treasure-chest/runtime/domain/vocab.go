package domain

// The seven string constants below are the ENTIRE surface of the
// strategist-skill internal/domain package that treasure-chest ever
// reached for. Copying just the vocabulary here severs the dependency
// without dragging in the ~10k LOC of internal/domain.
//
// If any of the strings change upstream, they change as coordinated
// contract updates — not silent drift.

// Confidence levels (upstream: internal/domain/decision.go).
const (
	ConfidenceLow    = "low"
	ConfidenceMedium = "medium"
	ConfidenceHigh   = "high"
)

// Evidence classification (upstream: internal/domain/evidence.go).
const (
	EvidenceClassExplicit              = "explicit"
	EvidenceClassCorroboratedInference = "corroborated_inference"
	EvidenceClassWeakInference         = "weak_inference"
	EvidenceClassUnknown               = "unknown"
)
