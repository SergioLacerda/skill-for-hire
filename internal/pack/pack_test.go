package pack_test

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SergioLacerda/skill-for-hire/internal/pack"
)

const testGenerator = "skillhire@test"
const testCommit = "0000000000000000000000000000000000000000"

func TestPackAtlasProducesArchiveAndChecksum(t *testing.T) {
	outDir := t.TempDir()
	res, err := pack.Pack(pack.Options{
		SkillDir:     "../../skills/atlas",
		SkillName:    "atlas",
		Version:      "0.1.0",
		OutDir:       outDir,
		Generator:    testGenerator,
		SourceCommit: testCommit,
	})
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	if res.Size == 0 {
		t.Fatal("archive is empty")
	}
	if res.SHA256 == "" {
		t.Fatal("sha256 missing")
	}

	if _, err := os.Stat(res.ArchivePath); err != nil {
		t.Fatalf("archive not written: %v", err)
	}
	if _, err := os.Stat(res.ChecksumPath); err != nil {
		t.Fatalf("checksum not written: %v", err)
	}

	names := listArchive(t, res.ArchivePath)
	// The archive must be rooted at the skill name, not at the source
	// path — consumers see atlas/SKILL.md, not skills/atlas/SKILL.md.
	mustContain(t, names, "atlas/SKILL.md")
	mustContain(t, names, "atlas/skill.yaml")
	mustContain(t, names, "atlas/references/capabilities.md")
	for _, n := range names {
		if strings.HasPrefix(n, "skills/") {
			t.Fatalf("archive leaks source layout: %s", n)
		}
	}
}

func TestPackDeterministicAcrossRuns(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	opts := pack.Options{
		SkillDir:     "../../skills/atlas",
		SkillName:    "atlas",
		Version:      "0.1.0",
		Generator:    testGenerator,
		SourceCommit: testCommit,
	}
	optsA := opts
	optsA.OutDir = dirA
	optsB := opts
	optsB.OutDir = dirB

	a, err := pack.Pack(optsA)
	if err != nil {
		t.Fatalf("pack A: %v", err)
	}
	b, err := pack.Pack(optsB)
	if err != nil {
		t.Fatalf("pack B: %v", err)
	}
	if a.SHA256 != b.SHA256 {
		t.Fatalf("expected identical digests, got %s vs %s", a.SHA256, b.SHA256)
	}
	// The sidecar manifest must be reproducible too — otherwise
	// digest-of-manifest checks by downstream consumers drift.
	mA := mustReadFile(t, a.ManifestPath)
	mB := mustReadFile(t, b.ManifestPath)
	if mA != mB {
		t.Fatalf("release manifest is not reproducible:\n--- A ---\n%s\n--- B ---\n%s", mA, mB)
	}
}

func TestPackWritesReleaseManifest(t *testing.T) {
	outDir := t.TempDir()
	res, err := pack.Pack(pack.Options{
		SkillDir:     "../../skills/atlas",
		SkillName:    "atlas",
		Version:      "0.1.0",
		OutDir:       outDir,
		Generator:    testGenerator,
		SourceCommit: testCommit,
	})
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	if res.ManifestPath == "" {
		t.Fatal("ManifestPath empty")
	}
	body := mustReadFile(t, res.ManifestPath)
	for _, want := range []string{
		"name: atlas",
		`version: "0.1.0"`,
		"archive: atlas-0.1.0.tar.gz",
		"digest: sha256:" + res.SHA256,
		"generator: " + testGenerator,
		"source_commit: " + testCommit,
		"atlas/SKILL.md",
		"atlas/skill.yaml",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("manifest missing %q\nfull:\n%s", want, body)
		}
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // path from test helper.
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func listArchive(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path) //nolint:gosec // test helper on a temp file.
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer func() { _ = f.Close() }()
	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer func() { _ = gr.Close() }()
	tr := tar.NewReader(gr)
	var out []string
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		out = append(out, filepath.ToSlash(hdr.Name))
	}
	return out
}

func mustContain(t *testing.T, names []string, want string) {
	t.Helper()
	for _, n := range names {
		if n == want {
			return
		}
	}
	t.Fatalf("archive missing entry %q; have: %v", want, names)
}
