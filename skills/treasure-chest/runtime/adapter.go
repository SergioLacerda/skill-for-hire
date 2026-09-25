package runtime

import (
	"context"
	"errors"
	"fmt"
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

// Search returns jewels and potions ranked by relevance to the query.
// Ranking policy lives in ranking.go and is deterministic across runs:
// same input query + same prepared index -> same result order.
//
// TokenBudget is applied AFTER ranking so consumers get the top N most
// relevant items rather than the first N encountered.
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
			score, matched := jewelMatch(j, needles)
			if !matched {
				continue
			}
			items = append(items, ka.Item{
				ID:            j.ID,
				Kind:          "jewel",
				Score:         score,
				Trust:         trustToFloat(j.Trust),
				Freshness:     ka.FreshnessFresh,
				Applicability: j.Status,
				Reason:        jewelReason(j, needles),
				Source: ka.Source{
					Path: firstSourceRef(j.SourceRefs),
				},
			})
		}
	}
	for _, list := range a.potionsByChest {
		for _, p := range list {
			score, matched := potionMatch(p, needles)
			if !matched {
				continue
			}
			items = append(items, ka.Item{
				ID:            p.ID,
				Kind:          "potion",
				Score:         score,
				Trust:         trustToFloat(p.Trust),
				Freshness:     ka.FreshnessFresh,
				Applicability: p.Status,
				Reason:        "when_to_use/runbook_ref matched query",
				Source: ka.Source{
					Path: p.RunbookRef,
				},
			})
		}
	}

	rankItemsInPlace(items)
	if q.TokenBudget > 0 && len(items) > q.TokenBudget {
		items = items[:q.TokenBudget]
	}

	return ka.SearchResult{
		Envelope: a.envelope(nil),
		Items:    items,
	}, nil
}

// jewelReason builds a compact, human-facing string explaining why a
// jewel matched. Keeps the ranking explainable (arch doc §5.5).
func jewelReason(j Jewel, needles []string) string {
	if len(needles) == 0 {
		return "no filter — matched by default"
	}
	// Cheap short-circuit: identify the strongest signal.
	for _, n := range needles {
		if containsFold(j.Statement, n) {
			return "statement matched query concept"
		}
		if anyContainsFold(j.Applicability.Scope, n) {
			return "applicability.scope matched query concept"
		}
		if anyContainsFold(j.Applicability.AppliesWhen, n) {
			return "applicability.applies_when matched query concept"
		}
		if containsFold(j.Kind, n) {
			return "kind matched query concept"
		}
	}
	return "curatorial score boost"
}

// Refresh reloads the index from the workspace and reports how it
// changed against the previous prepared state. The full reload is
// unavoidable until git-aware incremental refresh lands (see wave 4c
// TODO), but consumers still see accurate added/updated/removed counts.
//
// "updated" here means "same ID, different content signature". Content
// is signed by fmt.Sprintf on the Jewel/Potion struct — sufficient to
// detect changes without a hash dependency, and Go's %v produces a
// stable representation for the value types in this domain.
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

	oldSigs := a.snapshotSignatures()

	if _, err := a.Prepare(ctx, ka.PrepareRequest{Root: root}); err != nil {
		return ka.RefreshResult{}, err
	}

	added, updated, removed := diffSignatures(oldSigs, a.snapshotSignatures())

	limitations := []string{}
	if len(scope.Since) > 0 || len(scope.Paths) > 0 {
		limitations = append(limitations,
			"Refresh does full reindex; Scope.Since and Scope.Paths are ignored until git-aware refresh lands")
	}

	return ka.RefreshResult{
		Envelope: a.envelope(limitations),
		Added:    added,
		Updated:  updated,
		Removed:  removed,
	}, nil
}

// snapshotSignatures captures a content signature per item ID from the
// currently prepared index. Called before and after reload so we can
// diff without deep struct equality.
func (a *Adapter) snapshotSignatures() map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sig := make(map[string]string, len(a.byID))
	for id, j := range a.byID {
		sig[id] = fmt.Sprintf("%v", *j)
	}
	for _, list := range a.potionsByChest {
		for i := range list {
			p := &list[i]
			sig[p.ID] = fmt.Sprintf("%v", *p)
		}
	}
	return sig
}

// diffSignatures returns the added/updated/removed counts between two
// signature snapshots. Deterministic; tests rely on the exact triple.
func diffSignatures(oldSig, newSig map[string]string) (added, updated, removed int) {
	for id, s := range newSig {
		prev, ok := oldSig[id]
		switch {
		case !ok:
			added++
		case prev != s:
			updated++
		}
	}
	for id := range oldSig {
		if _, ok := newSig[id]; !ok {
			removed++
		}
	}
	return added, updated, removed
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
