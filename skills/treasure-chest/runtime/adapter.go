package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
)

// Adapter implements platform/knowledge-api.KnowledgeProvider on top of
// the treasure-chest domain. State is held in memory; call Prepare
// before Search/Refresh/Explain.
//
// Deliberately minimal: this first cut wires the existing loaders into
// the KA envelope so downstream consumers can hit the contract today.
// Ranking, freshness computation from git, and TTL live behind later
// batches.
type Adapter struct {
	// version is embedded into every response's Envelope.Provider.
	version string

	mu             sync.RWMutex
	root           string
	prepared       bool
	preparedAt     time.Time
	governed       map[string]GovernedChest
	jewelsByChest  map[string][]Jewel
	potionsByChest map[string][]Potion
	byID           map[string]*Jewel
}

// providerName is the fixed identity every Envelope carries. Consumers
// route by capability, not by name — this string is auditable metadata.
const providerName = "treasure-chest"

// capabilities is what this adapter implements. Must stay in sync with
// skills/treasure-chest/skill.yaml's spec.composition.provides.
var capabilities = []string{
	"knowledge.mine",
	"knowledge.search",
	"knowledge.curate",
	"runbook.select",
	"learning.reuse",
}

// NewAdapter returns a fresh adapter reporting the given version. Pass
// the skill's SemVer (e.g. "0.2.0") — the caller owns version drift, not
// this package.
func NewAdapter(version string) *Adapter {
	return &Adapter{version: version}
}

// Prepare loads governed chests, jewels and potions from req.Root and
// indexes them for Search/Explain. Idempotent — a second call reloads
// from disk, mirroring Refresh (until incremental refresh lands).
func (a *Adapter) Prepare(_ context.Context, req ka.PrepareRequest) (ka.PrepareResult, error) {
	if req.Root == "" {
		return ka.PrepareResult{}, fmt.Errorf("treasure-chest: PrepareRequest.Root is required")
	}

	governed, err := LoadGoverned(req.Root)
	if err != nil {
		return ka.PrepareResult{}, fmt.Errorf("treasure-chest: load governed: %w", err)
	}
	jewels, err := LoadJewels(req.Root, governed)
	if err != nil {
		return ka.PrepareResult{}, fmt.Errorf("treasure-chest: load jewels: %w", err)
	}
	potions, err := LoadPotions(req.Root, governed)
	if err != nil {
		return ka.PrepareResult{}, fmt.Errorf("treasure-chest: load potions: %w", err)
	}

	byID := make(map[string]*Jewel)
	sourcesRead := len(governed)
	itemsIndexed := 0
	for _, list := range jewels {
		for i := range list {
			byID[list[i].ID] = &list[i]
			itemsIndexed++
		}
	}
	for _, list := range potions {
		itemsIndexed += len(list)
	}

	a.mu.Lock()
	a.root = req.Root
	a.prepared = true
	a.preparedAt = time.Now().UTC()
	a.governed = governed
	a.jewelsByChest = jewels
	a.potionsByChest = potions
	a.byID = byID
	a.mu.Unlock()

	return ka.PrepareResult{
		Envelope:     a.envelope(nil),
		SourcesRead:  sourcesRead,
		ItemsIndexed: itemsIndexed,
	}, nil
}

// Search returns jewels and potions whose Statement contains any of the
// query concepts (case-insensitive substring). Naive on purpose — the
// domain already exposes richer filters, and richer ranking belongs in
// a later batch that wires the domain's scoring.
func (a *Adapter) Search(_ context.Context, q ka.Query) (ka.SearchResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.prepared {
		return ka.SearchResult{}, ka.ErrNotPrepared
	}

	needles := searchNeedles(q)
	items := make([]ka.Item, 0)

	for _, list := range a.jewelsByChest {
		for _, j := range list {
			if !matchesAny(j.Statement, needles) && !matchesAny(j.Kind, needles) {
				continue
			}
			items = append(items, ka.Item{
				ID:            j.ID,
				Kind:          "jewel",
				Score:         1.0,
				Freshness:     ka.FreshnessFresh,
				Applicability: j.Status,
				Reason:        "statement or kind matched query concept",
				Source: ka.Source{
					Path:   firstSourceRef(j.SourceRefs),
					Digest: "",
				},
			})
			if q.TokenBudget > 0 && len(items) >= q.TokenBudget {
				break
			}
		}
	}
	for _, list := range a.potionsByChest {
		for _, p := range list {
			if !matchesAny(p.WhenToUse, needles) && !matchesAny(p.RunbookRef, needles) {
				continue
			}
			items = append(items, ka.Item{
				ID:            p.ID,
				Kind:          "potion",
				Score:         0.9,
				Freshness:     ka.FreshnessFresh,
				Applicability: p.Status,
				Reason:        "when_to_use or runbook_ref matched query concept",
				Source: ka.Source{
					Path: p.RunbookRef,
				},
			})
			if q.TokenBudget > 0 && len(items) >= q.TokenBudget {
				break
			}
		}
	}

	return ka.SearchResult{
		Envelope: a.envelope(nil),
		Items:    items,
	}, nil
}

// Refresh currently reloads from the same root — the domain does not
// yet expose a scoped diff. Callers see Added=0/Updated=0/Removed=0
// until incremental refresh lands.
func (a *Adapter) Refresh(ctx context.Context, scope ka.Scope) (ka.RefreshResult, error) {
	root := scope.Root
	if root == "" {
		a.mu.RLock()
		root = a.root
		a.mu.RUnlock()
	}
	if root == "" {
		return ka.RefreshResult{}, ka.ErrNotPrepared
	}
	if _, err := a.Prepare(ctx, ka.PrepareRequest{Root: root}); err != nil {
		return ka.RefreshResult{}, err
	}
	return ka.RefreshResult{
		Envelope: a.envelope([]string{"refresh does full reindex; incremental diff pending"}),
	}, nil
}

// Status reports provider identity plus item counts. Consumers that
// route by capability use this call before dispatching work.
func (a *Adapter) Status(_ context.Context) (ka.Status, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	detail := "not prepared"
	if a.prepared {
		items := 0
		for _, l := range a.jewelsByChest {
			items += len(l)
		}
		for _, l := range a.potionsByChest {
			items += len(l)
		}
		detail = fmt.Sprintf("%d chests, %d items indexed at %s",
			len(a.governed), items, a.preparedAt.Format(time.RFC3339))
	}
	return ka.Status{
		Provider:      providerName,
		Version:       a.version,
		SchemaVersion: ka.SchemaVersion,
		Healthy:       a.prepared,
		Capabilities:  capabilities,
		LastPrepareAt: a.preparedAt.Format(time.RFC3339),
		Detail:        detail,
	}, nil
}

// Explain looks up a jewel by ID and reports the routing decision. If
// the ID is unknown after Prepare, the decision is Discarded with the
// reason recorded on the response.
func (a *Adapter) Explain(_ context.Context, itemID string) (ka.Explanation, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.prepared {
		return ka.Explanation{}, ka.ErrNotPrepared
	}
	j, ok := a.byID[itemID]
	if !ok {
		return ka.Explanation{
			ItemID:   itemID,
			Decision: ka.DecisionDiscarded,
			Reason:   "id not present in current index",
		}, nil
	}
	sources := make([]ka.Source, 0, len(j.SourceRefs))
	for _, ref := range j.SourceRefs {
		sources = append(sources, ka.Source{Path: ref})
	}
	return ka.Explanation{
		ItemID:   itemID,
		Decision: ka.DecisionSelected,
		Reason:   fmt.Sprintf("jewel %q in chest %q, status=%s, trust=%s", j.Kind, j.ChestID, j.Status, j.Trust),
		Evidence: sources,
	}, nil
}

// envelope builds the response envelope consumers get on every call.
func (a *Adapter) envelope(limitations []string) ka.Envelope {
	freshness := ka.FreshnessUnknown
	if a.prepared {
		freshness = ka.FreshnessFresh
	}
	return ka.Envelope{
		Provider:         providerName,
		Version:          a.version,
		SchemaVersion:    ka.SchemaVersion,
		CapabilitiesUsed: capabilities,
		Freshness:        freshness,
		Trust:            1.0,
		FallbackState:    ka.FallbackNone,
		Limitations:      limitations,
	}
}

// Compile-time proof that Adapter satisfies KnowledgeProvider.
var _ ka.KnowledgeProvider = (*Adapter)(nil)

// searchNeedles collects the substrings Search matches against.
func searchNeedles(q ka.Query) []string {
	out := make([]string, 0, len(q.Concepts)+len(q.Symbols)+1)
	if q.Intent != "" {
		out = append(out, q.Intent)
	}
	out = append(out, q.Concepts...)
	out = append(out, q.Symbols...)
	return out
}

func matchesAny(haystack string, needles []string) bool {
	if len(needles) == 0 {
		return true // no filter = match all
	}
	hay := strings.ToLower(haystack)
	for _, n := range needles {
		if n == "" {
			continue
		}
		if strings.Contains(hay, strings.ToLower(n)) {
			return true
		}
	}
	return false
}

func firstSourceRef(refs []string) string {
	if len(refs) == 0 {
		return ""
	}
	return refs[0]
}

// Guard against a linker footgun: if capabilities and skill.yaml drift
// apart, at least the code path checks itself.
var errCapabilitiesEmpty = errors.New("treasure-chest: capabilities list must not be empty")

func init() {
	if len(capabilities) == 0 {
		panic(errCapabilitiesEmpty)
	}
}
