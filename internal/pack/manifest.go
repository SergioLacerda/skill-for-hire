package pack

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// manifest is the sidecar written next to <name>-<version>.tar.gz as
// <name>-<version>.release.yaml. It matches the shape described in
// docs/architecture, §5.5 Packaging: name, version, digest, generator,
// source commit, and the archive-relative content list. Deliberately
// hand-serialized (no yaml dep here) so the file is byte-identical
// across runs and reproducibility is provable.
type manifest struct {
	Name          string
	Version       string
	SchemaVersion int
	Digest        string
	Size          int64
	Generator     string
	SourceCommit  string
	Archive       string
	Contents      []string
}

func writeManifest(path string, m manifest) error {
	body := renderManifest(m)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil { //nolint:gosec // 0644 is standard for a public manifest file.
		return fmt.Errorf("pack: write manifest: %w", err)
	}
	return nil
}

func renderManifest(m manifest) string {
	// Sort the content list so the manifest bytes do not depend on walk
	// order. collectEntries already sorts, but this is cheap belt-and-
	// suspenders and keeps the manifest independent of the archive's
	// internal ordering.
	sorted := make([]string, len(m.Contents))
	copy(sorted, m.Contents)
	sort.Strings(sorted)

	var b strings.Builder
	b.WriteString("schema_version: ")
	fmt.Fprintf(&b, "%d\n", m.SchemaVersion)
	fmt.Fprintf(&b, "name: %s\n", m.Name)
	fmt.Fprintf(&b, "version: %q\n", m.Version)
	fmt.Fprintf(&b, "archive: %s\n", m.Archive)
	fmt.Fprintf(&b, "digest: %s\n", m.Digest)
	fmt.Fprintf(&b, "size: %d\n", m.Size)
	if m.Generator != "" {
		fmt.Fprintf(&b, "generator: %s\n", m.Generator)
	}
	if m.SourceCommit != "" {
		fmt.Fprintf(&b, "source_commit: %s\n", m.SourceCommit)
	}
	b.WriteString("contents:\n")
	for _, c := range sorted {
		fmt.Fprintf(&b, "  - %s\n", c)
	}
	return b.String()
}
