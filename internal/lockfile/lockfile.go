// Package lockfile reads and writes skillhire.lock — a consumer-side pin
// of installed skills to exact versions and digests. Generation is
// deterministic (sorted keys, no timestamps) so a lockfile committed to
// a repo has stable diffs.
package lockfile

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// SchemaVersion is the only accepted schema-version for now.
const SchemaVersion = 1

// Lockfile is the top-level structure of skillhire.lock.
type Lockfile struct {
	SchemaVersion int              `yaml:"schema-version"`
	Generator     string           `yaml:"generator,omitempty"`
	Installed     map[string]Entry `yaml:"installed"`
}

// Entry pins one skill.
type Entry struct {
	Version      string `yaml:"version"`
	Digest       string `yaml:"digest"`
	Size         int64  `yaml:"size,omitempty"`
	Source       string `yaml:"source"`
	ResolvedFrom string `yaml:"resolved-from,omitempty"`
}

// Load reads and decodes a lockfile.
func Load(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from a CLI arg.
	if err != nil {
		return nil, fmt.Errorf("lockfile: read %s: %w", path, err)
	}
	var lf Lockfile
	if err := yaml.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("lockfile: parse %s: %w", path, err)
	}
	if lf.SchemaVersion == 0 {
		return nil, fmt.Errorf("lockfile: missing schema-version in %s", path)
	}
	if lf.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("lockfile: unsupported schema-version %d (want %d)", lf.SchemaVersion, SchemaVersion)
	}
	return &lf, nil
}

// Write serializes the lockfile deterministically. Keys are sorted so the
// on-disk bytes only change when the pinned content changes.
func Write(path string, lf *Lockfile) error {
	body, err := Render(lf)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil { //nolint:gosec // 0644 is standard for a public lockfile.
		return fmt.Errorf("lockfile: write %s: %w", path, err)
	}
	return nil
}

// Render returns the YAML body Write would emit. Split out so tests can
// diff strings without hitting disk.
func Render(lf *Lockfile) (string, error) {
	if lf.SchemaVersion == 0 {
		lf.SchemaVersion = SchemaVersion
	}
	// yaml.v3 preserves map insertion order via yaml.Node; we build the
	// document explicitly to keep the top-level and inner keys sorted.
	root := &yaml.Node{Kind: yaml.MappingNode}
	appendScalar(root, "schema-version", fmt.Sprintf("%d", lf.SchemaVersion), false)
	if lf.Generator != "" {
		appendScalar(root, "generator", lf.Generator, true)
	}

	installed := &yaml.Node{Kind: yaml.MappingNode}
	appendKey(root, "installed", installed)

	names := make([]string, 0, len(lf.Installed))
	for name := range lf.Installed {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entry := lf.Installed[name]
		node := entryNode(entry)
		appendKey(installed, name, node)
	}

	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		return "", fmt.Errorf("lockfile: encode: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", fmt.Errorf("lockfile: close encoder: %w", err)
	}
	return buf.String(), nil
}

// DiffInstalled reports a human-readable diff of the pinned entries in
// two lockfiles, ignoring the generator field. An empty string means
// the pinned state is identical.
func DiffInstalled(a, b *Lockfile) string {
	names := make(map[string]struct{}, len(a.Installed)+len(b.Installed))
	for name := range a.Installed {
		names[name] = struct{}{}
	}
	for name := range b.Installed {
		names[name] = struct{}{}
	}
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)

	var lines []string
	for _, name := range sorted {
		lhs, hasA := a.Installed[name]
		rhs, hasB := b.Installed[name]
		switch {
		case !hasA:
			lines = append(lines, fmt.Sprintf("+ %s: added (version=%s digest=%s)", name, rhs.Version, rhs.Digest))
		case !hasB:
			lines = append(lines, fmt.Sprintf("- %s: removed", name))
		case lhs != rhs:
			lines = append(lines, fmt.Sprintf("~ %s: %s@%s -> %s@%s", name, lhs.Version, lhs.Digest, rhs.Version, rhs.Digest))
		}
	}
	return strings.Join(lines, "\n")
}

// FromManifests builds a lockfile from a set of release.yaml sidecars.
// Path/source pairs come from the caller so this package has no opinion
// about where the artifacts live (monorepo vs github-release vs mirror).
func FromManifests(inputs []ManifestInput, generator string) (*Lockfile, error) {
	lf := &Lockfile{
		SchemaVersion: SchemaVersion,
		Generator:     generator,
		Installed:     make(map[string]Entry, len(inputs)),
	}
	for _, in := range inputs {
		m, err := loadManifest(in.Path)
		if err != nil {
			return nil, err
		}
		if _, dup := lf.Installed[m.Name]; dup {
			return nil, fmt.Errorf("lockfile: duplicate skill %q across manifests", m.Name)
		}
		source := in.Source
		if source == "" {
			// Default derives from the manifest's own name so callers
			// don't have to parse basenames like "treasure-chest-…".
			source = "monorepo:skills/" + m.Name
		}
		lf.Installed[m.Name] = Entry{
			Version:      m.Version,
			Digest:       m.Digest,
			Size:         m.Size,
			Source:       source,
			ResolvedFrom: in.ResolvedFrom,
		}
	}
	return lf, nil
}

// ManifestInput names one release.yaml plus the source string and
// resolved-from string the caller wants recorded.
type ManifestInput struct {
	Path         string
	Source       string
	ResolvedFrom string
}

// releaseManifest mirrors the sidecar written by internal/pack. Kept
// tiny — only the fields the lockfile needs.
type releaseManifest struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Digest  string `yaml:"digest"`
	Size    int64  `yaml:"size"`
}

func loadManifest(path string) (*releaseManifest, error) {
	data, err := os.ReadFile(path) //nolint:gosec // vetted CLI input.
	if err != nil {
		return nil, fmt.Errorf("lockfile: read manifest %s: %w", path, err)
	}
	var m releaseManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("lockfile: parse manifest %s: %w", path, err)
	}
	if m.Name == "" || m.Version == "" || m.Digest == "" {
		return nil, fmt.Errorf("lockfile: manifest %s missing name/version/digest", path)
	}
	return &m, nil
}

// entryNode builds a mapping node with the entry's fields in a fixed
// order: version, digest, size, source, resolved-from.
func entryNode(e Entry) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	appendScalar(n, "version", e.Version, true)
	appendScalar(n, "digest", e.Digest, false)
	if e.Size > 0 {
		appendScalar(n, "size", fmt.Sprintf("%d", e.Size), false)
	}
	appendScalar(n, "source", e.Source, false)
	if e.ResolvedFrom != "" {
		appendScalar(n, "resolved-from", e.ResolvedFrom, false)
	}
	return n
}

func appendScalar(parent *yaml.Node, key, value string, quoteValue bool) {
	style := yaml.Style(0)
	if quoteValue {
		style = yaml.DoubleQuotedStyle
	}
	parent.Content = append(parent.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value, Style: style},
	)
}

func appendKey(parent *yaml.Node, key string, value *yaml.Node) {
	parent.Content = append(parent.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		value,
	)
}
