package lockfile_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SergioLacerda/skill-for-hire/internal/lockfile"
)

const sampleManifestA = `schema_version: 1
name: atlas
version: "0.1.0"
archive: atlas-0.1.0.tar.gz
digest: sha256:` + digestA + `
size: 4514
generator: skillhire@test
source_commit: 0000000000000000000000000000000000000000
contents:
  - atlas/SKILL.md
  - atlas/skill.yaml
`

const sampleManifestB = `schema_version: 1
name: treasure-chest
version: "0.1.0"
archive: treasure-chest-0.1.0.tar.gz
digest: sha256:` + digestB + `
size: 4306
generator: skillhire@test
contents:
  - treasure-chest/SKILL.md
  - treasure-chest/skill.yaml
`

const (
	digestA = "f8671ffdec326ca0000998321e949775ce6afc7f4fe6a0e881c21ab410329f02"
	digestB = "03eebd0e91e81c9ee757287913c75971f7b0504975ca08e544d30d962e15aad3"
)

func TestFromManifestsAndRenderIsSortedAndDeterministic(t *testing.T) {
	tmp := t.TempDir()
	pathA := writeFile(t, tmp, "atlas-0.1.0.release.yaml", sampleManifestA)
	pathB := writeFile(t, tmp, "treasure-chest-0.1.0.release.yaml", sampleManifestB)

	// Feed B before A; render must still sort skills alphabetically.
	// Source left empty on both to exercise the "default from manifest
	// name" path in FromManifests.
	lf, err := lockfile.FromManifests([]lockfile.ManifestInput{
		{Path: pathB, ResolvedFrom: "treasure-chest-0.1.0.release.yaml"},
		{Path: pathA, ResolvedFrom: "atlas-0.1.0.release.yaml"},
	}, "skillhire@test")
	if err != nil {
		t.Fatalf("FromManifests: %v", err)
	}

	rendered, err := lockfile.Render(lf)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// atlas comes before treasure-chest.
	if idxA := strings.Index(rendered, "atlas:"); idxA < 0 {
		t.Fatalf("rendered lockfile missing atlas:\n%s", rendered)
	} else if idxB := strings.Index(rendered, "treasure-chest:"); idxB < 0 {
		t.Fatalf("rendered lockfile missing treasure-chest:\n%s", rendered)
	} else if idxA >= idxB {
		t.Fatalf("skills not sorted: atlas at %d, treasure-chest at %d\n%s", idxA, idxB, rendered)
	}

	// Determinism: encode twice, bytes must match.
	rendered2, err := lockfile.Render(lf)
	if err != nil {
		t.Fatalf("Render (2): %v", err)
	}
	if rendered != rendered2 {
		t.Fatalf("Render not deterministic:\n--- 1 ---\n%s\n--- 2 ---\n%s", rendered, rendered2)
	}

	for _, want := range []string{
		"schema-version: 1",
		`generator: "skillhire@test"`,
		"atlas:",
		`version: "0.1.0"`,
		"digest: sha256:" + digestA,
		"source: monorepo:skills/atlas",
		"resolved-from: atlas-0.1.0.release.yaml",
		"treasure-chest:",
		"source: monorepo:skills/treasure-chest",
		"digest: sha256:" + digestB,
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered lockfile missing %q\nfull:\n%s", want, rendered)
		}
	}
}

func TestLoadRejectsUnknownSchemaVersion(t *testing.T) {
	tmp := t.TempDir()
	p := writeFile(t, tmp, "skillhire.lock", "schema-version: 999\ninstalled: {}\n")
	if _, err := lockfile.Load(p); err == nil {
		t.Fatal("expected error for unsupported schema-version, got nil")
	}
}

func TestFromManifestsDetectsDuplicates(t *testing.T) {
	tmp := t.TempDir()
	pathA := writeFile(t, tmp, "atlas-0.1.0.release.yaml", sampleManifestA)
	pathAcopy := writeFile(t, tmp, "atlas-copy.release.yaml", sampleManifestA)
	_, err := lockfile.FromManifests([]lockfile.ManifestInput{
		{Path: pathA, Source: "monorepo:skills/atlas"},
		{Path: pathAcopy, Source: "mirror:acme/atlas"},
	}, "skillhire@test")
	if err == nil {
		t.Fatal("expected duplicate skill error, got nil")
	}
}

func TestRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	pathA := writeFile(t, tmp, "atlas-0.1.0.release.yaml", sampleManifestA)
	lf, err := lockfile.FromManifests([]lockfile.ManifestInput{
		{Path: pathA, Source: "monorepo:skills/atlas"},
	}, "skillhire@test")
	if err != nil {
		t.Fatalf("FromManifests: %v", err)
	}
	lockPath := filepath.Join(tmp, "skillhire.lock")
	if err := lockfile.Write(lockPath, lf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	reloaded, err := lockfile.Load(lockPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry, ok := reloaded.Installed["atlas"]
	if !ok {
		t.Fatal("reloaded lockfile missing atlas")
	}
	if entry.Digest != "sha256:"+digestA {
		t.Fatalf("digest drift: %s", entry.Digest)
	}
	if entry.Version != "0.1.0" {
		t.Fatalf("version drift: %s", entry.Version)
	}
}

func writeFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil { //nolint:gosec // test helper.
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}
