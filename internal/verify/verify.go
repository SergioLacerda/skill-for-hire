// Package verify checks a downloaded skill pack against its digest and
// (when present) its release manifest sidecar. Consumers use this after
// downloading a release asset, before extracting or trusting the pack.
package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Result reports what Verify checked and what was consistent.
type Result struct {
	ArchivePath    string
	ChecksumPath   string
	ManifestPath   string // empty when no sidecar found
	SHA256         string
	Size           int64
	ManifestExists bool
}

// Verify runs the consumer-side checks:
//
//  1. Compute sha256 of the archive.
//  2. Read the paired <archive>.sha256 companion and compare.
//  3. If <name>-<version>.release.yaml sits next to the archive, cross-
//     check its `digest` and `size` fields against what we just measured.
//
// A missing manifest is not an error (older packs may not ship one); a
// present-but-mismatched manifest IS an error, since it signals tampering
// between digest and metadata.
func Verify(archivePath string) (*Result, error) {
	result := &Result{
		ArchivePath: archivePath,
	}

	info, err := os.Stat(archivePath)
	if err != nil {
		return nil, fmt.Errorf("verify: stat archive: %w", err)
	}
	result.Size = info.Size()

	digest, err := digestFile(archivePath)
	if err != nil {
		return nil, err
	}
	result.SHA256 = digest

	checksumPath := archivePath + ".sha256"
	result.ChecksumPath = checksumPath
	expected, err := readChecksumCompanion(checksumPath, filepath.Base(archivePath))
	if err != nil {
		return nil, err
	}
	if expected != digest {
		return nil, fmt.Errorf("verify: sha256 mismatch for %s\n  expected: %s (from %s)\n  actual:   %s",
			archivePath, expected, checksumPath, digest)
	}

	manifestPath := deriveManifestPath(archivePath)
	if manifestPath == "" {
		return result, nil
	}
	if _, err := os.Stat(manifestPath); errors.Is(err, os.ErrNotExist) {
		return result, nil
	} else if err != nil {
		return nil, fmt.Errorf("verify: stat manifest: %w", err)
	}
	result.ManifestPath = manifestPath
	result.ManifestExists = true

	if err := crossCheckManifest(manifestPath, digest, result.Size); err != nil {
		return nil, err
	}
	return result, nil
}

// deriveManifestPath turns <name>-<version>.tar.gz into
// <name>-<version>.release.yaml. Returns "" when the archive name does
// not end with .tar.gz.
func deriveManifestPath(archivePath string) string {
	base := filepath.Base(archivePath)
	if !strings.HasSuffix(base, ".tar.gz") {
		return ""
	}
	stem := strings.TrimSuffix(base, ".tar.gz")
	return filepath.Join(filepath.Dir(archivePath), stem+".release.yaml")
}

func digestFile(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // caller-supplied archive path.
	if err != nil {
		return "", fmt.Errorf("verify: open archive: %w", err)
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("verify: read archive: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// readChecksumCompanion accepts either the coreutils format
// "digest  filename" or a bare digest line. The filename must match
// the archive basename when present, so a mixed-up .sha256 does not
// silently pass.
func readChecksumCompanion(path, wantBasename string) (string, error) {
	body, err := os.ReadFile(path) //nolint:gosec // caller-supplied path.
	if err != nil {
		return "", fmt.Errorf("verify: read checksum: %w", err)
	}
	line := strings.TrimSpace(string(body))
	if line == "" {
		return "", fmt.Errorf("verify: empty checksum file %s", path)
	}
	fields := strings.Fields(line)
	digest := fields[0]
	if len(fields) >= 2 {
		claimed := fields[len(fields)-1]
		if claimed != wantBasename {
			return "", fmt.Errorf("verify: checksum file %s references %q but archive basename is %q",
				path, claimed, wantBasename)
		}
	}
	if len(digest) != 64 {
		return "", fmt.Errorf("verify: %s does not look like a sha256 hex digest", digest)
	}
	return digest, nil
}

// crossCheckManifest reads a release.yaml and validates that its
// digest/size match what we just measured on the archive.
func crossCheckManifest(path, digest string, size int64) error {
	body, err := os.ReadFile(path) //nolint:gosec // sibling of caller-supplied archive.
	if err != nil {
		return fmt.Errorf("verify: read manifest: %w", err)
	}
	var m struct {
		Digest string `yaml:"digest"`
		Size   int64  `yaml:"size"`
	}
	if err := yaml.Unmarshal(body, &m); err != nil {
		return fmt.Errorf("verify: parse manifest: %w", err)
	}
	want := "sha256:" + digest
	if m.Digest != "" && m.Digest != want {
		return fmt.Errorf("verify: manifest digest disagrees with archive\n  manifest: %s\n  archive:  %s", m.Digest, want)
	}
	if m.Size != 0 && m.Size != size {
		return fmt.Errorf("verify: manifest size %d disagrees with archive size %d", m.Size, size)
	}
	return nil
}
