package verify_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SergioLacerda/skill-for-hire/internal/pack"
	"github.com/SergioLacerda/skill-for-hire/internal/verify"
)

func TestVerifyAcceptsFreshPack(t *testing.T) {
	outDir := t.TempDir()
	res, err := pack.Pack(pack.Options{
		SkillDir:     "../../skills/atlas",
		SkillName:    "atlas",
		Version:      "0.1.0",
		OutDir:       outDir,
		Generator:    "skillhire@test",
		SourceCommit: "0000000000000000000000000000000000000000",
	})
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	r, err := verify.Verify(res.ArchivePath)
	if err != nil {
		t.Fatalf("verify failed on a fresh pack: %v", err)
	}
	if r.SHA256 != res.SHA256 {
		t.Fatalf("digest drift: %s vs %s", r.SHA256, res.SHA256)
	}
	if !r.ManifestExists {
		t.Fatal("verify should have found the release.yaml sidecar")
	}
}

func TestVerifyRejectsTamperedArchive(t *testing.T) {
	outDir := t.TempDir()
	res, err := pack.Pack(pack.Options{
		SkillDir:     "../../skills/atlas",
		SkillName:    "atlas",
		Version:      "0.1.0",
		OutDir:       outDir,
		Generator:    "skillhire@test",
		SourceCommit: "0000000000000000000000000000000000000000",
	})
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	// Append a byte to the archive without updating the .sha256.
	f, err := os.OpenFile(res.ArchivePath, os.O_APPEND|os.O_WRONLY, 0o644) //nolint:gosec // temp file.
	if err != nil {
		t.Fatalf("open for append: %v", err)
	}
	if _, err := f.Write([]byte{0}); err != nil {
		t.Fatalf("append: %v", err)
	}
	_ = f.Close()

	_, err = verify.Verify(res.ArchivePath)
	if err == nil {
		t.Fatal("expected sha256 mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "sha256 mismatch") {
		t.Fatalf("expected sha256 mismatch error, got: %v", err)
	}
}

func TestVerifyRejectsTamperedManifest(t *testing.T) {
	outDir := t.TempDir()
	res, err := pack.Pack(pack.Options{
		SkillDir:     "../../skills/atlas",
		SkillName:    "atlas",
		Version:      "0.1.0",
		OutDir:       outDir,
		Generator:    "skillhire@test",
		SourceCommit: "0000000000000000000000000000000000000000",
	})
	if err != nil {
		t.Fatalf("pack: %v", err)
	}

	// Rewrite the manifest digest to something wrong; archive + .sha256
	// stay consistent with each other.
	body, err := os.ReadFile(res.ManifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	tampered := strings.Replace(string(body),
		"digest: sha256:"+res.SHA256,
		"digest: sha256:"+strings.Repeat("0", 64),
		1)
	if tampered == string(body) {
		t.Fatal("test setup: manifest digest not replaced")
	}
	if err := os.WriteFile(res.ManifestPath, []byte(tampered), 0o644); err != nil { //nolint:gosec // temp file.
		t.Fatalf("write manifest: %v", err)
	}

	_, err = verify.Verify(res.ArchivePath)
	if err == nil {
		t.Fatal("expected manifest-vs-archive mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "manifest digest disagrees") {
		t.Fatalf("expected manifest-disagrees error, got: %v", err)
	}
}

func TestVerifyRejectsChecksumFileNamingWrongFile(t *testing.T) {
	outDir := t.TempDir()
	res, err := pack.Pack(pack.Options{
		SkillDir:     "../../skills/atlas",
		SkillName:    "atlas",
		Version:      "0.1.0",
		OutDir:       outDir,
		Generator:    "skillhire@test",
		SourceCommit: "0000000000000000000000000000000000000000",
	})
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	// Point the .sha256 at a different filename — even if the digest
	// happens to match, this is a supply-chain smell.
	spoof := res.SHA256 + "  someone-elses-archive.tar.gz\n"
	if err := os.WriteFile(res.ChecksumPath, []byte(spoof), 0o644); err != nil { //nolint:gosec // temp file.
		t.Fatalf("spoof checksum: %v", err)
	}
	_, err = verify.Verify(res.ArchivePath)
	if err == nil {
		t.Fatal("expected checksum-references-wrong-basename error")
	}
	if !strings.Contains(err.Error(), "references") {
		t.Fatalf("expected 'references' error, got: %v", err)
	}
}

func TestVerifyAcceptsBareDigestChecksum(t *testing.T) {
	outDir := t.TempDir()
	res, err := pack.Pack(pack.Options{
		SkillDir:     "../../skills/atlas",
		SkillName:    "atlas",
		Version:      "0.1.0",
		OutDir:       outDir,
		Generator:    "skillhire@test",
		SourceCommit: "0000000000000000000000000000000000000000",
	})
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	// Some publishers ship just the hex digest; still acceptable.
	if err := os.WriteFile(res.ChecksumPath, []byte(res.SHA256+"\n"), 0o644); err != nil { //nolint:gosec // temp file.
		t.Fatalf("bare digest: %v", err)
	}
	if _, err := verify.Verify(res.ArchivePath); err != nil {
		t.Fatalf("verify bare digest: %v", err)
	}
}

func TestVerifyToleratesMissingManifest(t *testing.T) {
	outDir := t.TempDir()
	res, err := pack.Pack(pack.Options{
		SkillDir:     "../../skills/atlas",
		SkillName:    "atlas",
		Version:      "0.1.0",
		OutDir:       outDir,
		Generator:    "skillhire@test",
		SourceCommit: "0000000000000000000000000000000000000000",
	})
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	if err := os.Remove(res.ManifestPath); err != nil {
		t.Fatalf("remove manifest: %v", err)
	}
	r, err := verify.Verify(res.ArchivePath)
	if err != nil {
		t.Fatalf("verify without manifest: %v", err)
	}
	if r.ManifestExists {
		t.Fatal("ManifestExists should be false when the sidecar is absent")
	}
}

// keeps t.TempDir usage compact.
var _ = filepath.Join
