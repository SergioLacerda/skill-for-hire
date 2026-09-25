// Package pack produces a reproducible tar.gz of a skill package plus a
// SHA-256 companion. The output layout follows ORKA: the archive root is
// the skill name, not the source directory, so the on-disk layout of the
// monorepo does not leak into consumers.
package pack

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Options controls Pack.
type Options struct {
	// SkillDir is the source directory (e.g. skills/atlas).
	SkillDir string
	// SkillName is the archive root and the filename stem.
	SkillName string
	// Version becomes part of the filename: <name>-<version>.tar.gz.
	Version string
	// OutDir is where the archive and .sha256 land.
	OutDir string
	// Include restricts what goes into the archive. If empty, defaults to
	// the ORKA canonical set.
	Include []string
	// IncludeTests keeps *_test.go files. Default false: consumers get a
	// slim runtime pack; contributors clone the repo.
	IncludeTests bool
	// Timestamp is written into every tar header for reproducibility. If
	// zero, the epoch is used.
	Timestamp time.Time
	// Generator identifies the tool that produced the archive. Ends up in
	// the sidecar manifest so consumers can trace bundle provenance.
	Generator string
	// SourceCommit is the git sha the pack was built from, when known.
	SourceCommit string
}

// Result reports what Pack produced.
type Result struct {
	ArchivePath  string
	ChecksumPath string
	ManifestPath string
	SHA256       string
	Size         int64
	Contents     []string
}

// defaultInclude is the ORKA canonical top-level set plus a few
// Skills-for-Hire extensions used by skills that ship runtime code.
// Each entry is tried as a file first and, failing that, a directory
// tree. Missing entries are silently skipped so a hello-world skill
// without runtime/ still packs cleanly.
var defaultInclude = []string{
	"SKILL.md",
	"skill.yaml",
	"README.md",
	"CHANGELOG.md",
	"references",
	"scripts",
	"templates",
	"assets",
	"runtime",   // Skills-for-Hire extension: skill-side Go/other runtime.
	"contracts", // Skills-for-Hire extension: contracts/schemas.
}

// Pack builds the archive and returns the paths and digest.
func Pack(opts Options) (*Result, error) {
	if opts.SkillDir == "" || opts.SkillName == "" || opts.Version == "" || opts.OutDir == "" {
		return nil, errors.New("pack: SkillDir, SkillName, Version and OutDir are required")
	}
	include := opts.Include
	if len(include) == 0 {
		include = defaultInclude
	}
	if opts.Timestamp.IsZero() {
		opts.Timestamp = time.Unix(0, 0).UTC()
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, fmt.Errorf("pack: create out dir: %w", err)
	}

	archivePath := filepath.Join(opts.OutDir, fmt.Sprintf("%s-%s.tar.gz", opts.SkillName, opts.Version))
	f, err := os.Create(archivePath) //nolint:gosec // path built from vetted inputs above.
	if err != nil {
		return nil, fmt.Errorf("pack: create archive: %w", err)
	}
	// Hash and write concurrently via a MultiWriter — one pass over the
	// bytes, no temporary re-read.
	hasher := sha256.New()
	gw := gzip.NewWriter(io.MultiWriter(f, hasher))
	tw := tar.NewWriter(gw)

	entries, err := collectEntries(opts.SkillDir, include)
	if err != nil {
		_ = tw.Close()
		_ = gw.Close()
		_ = f.Close()
		return nil, err
	}
	if !opts.IncludeTests {
		entries = filterOutTests(entries)
	}
	for _, e := range entries {
		if err := writeEntry(tw, opts.SkillName, opts.SkillDir, e, opts.Timestamp); err != nil {
			_ = tw.Close()
			_ = gw.Close()
			_ = f.Close()
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		_ = gw.Close()
		_ = f.Close()
		return nil, fmt.Errorf("pack: close tar: %w", err)
	}
	if err := gw.Close(); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("pack: close gzip: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("pack: close archive: %w", err)
	}

	size, err := fileSize(archivePath)
	if err != nil {
		return nil, err
	}
	digest := hex.EncodeToString(hasher.Sum(nil))
	checksumPath := archivePath + ".sha256"
	line := fmt.Sprintf("%s  %s\n", digest, filepath.Base(archivePath))
	if err := os.WriteFile(checksumPath, []byte(line), 0o644); err != nil { //nolint:gosec // 0644 is standard for a public digest file.
		return nil, fmt.Errorf("pack: write checksum: %w", err)
	}

	manifestPath := filepath.Join(opts.OutDir, fmt.Sprintf("%s-%s.release.yaml", opts.SkillName, opts.Version))
	contents := archiveContents(opts.SkillName, entries)
	if err := writeManifest(manifestPath, manifest{
		Name:          opts.SkillName,
		Version:       opts.Version,
		SchemaVersion: 1,
		Digest:        "sha256:" + digest,
		Size:          size,
		Generator:     opts.Generator,
		SourceCommit:  opts.SourceCommit,
		Archive:       filepath.Base(archivePath),
		Contents:      contents,
	}); err != nil {
		return nil, err
	}
	return &Result{
		ArchivePath:  archivePath,
		ChecksumPath: checksumPath,
		ManifestPath: manifestPath,
		SHA256:       digest,
		Size:         size,
		Contents:     contents,
	}, nil
}

// archiveContents returns the archive-relative paths that made it into
// the tarball, matching what an extractor sees.
func archiveContents(archiveRoot string, entries []string) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = archiveRoot + "/" + e
	}
	return out
}

func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("pack: stat %s: %w", path, err)
	}
	return info.Size(), nil
}

// filterOutTests drops Go test files from the pack. Consumers don't run
// them (they need the module context) and dropping them keeps the pack
// small — treasure-chest goes from ~68KB to <20KB, atlas is unchanged.
func filterOutTests(entries []string) []string {
	out := entries[:0]
	for _, e := range entries {
		if strings.HasSuffix(e, "_test.go") {
			continue
		}
		out = append(out, e)
	}
	return out
}

// collectEntries walks the include set and returns file relative paths in
// deterministic order. Directory entries are emitted before their
// children so extractors keep permissions.
func collectEntries(root string, include []string) ([]string, error) {
	seen := make(map[string]struct{})
	var out []string
	for _, top := range include {
		abs := filepath.Join(root, top)
		info, err := os.Stat(abs)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue // optional element absent — that is fine.
			}
			return nil, fmt.Errorf("pack: stat %s: %w", abs, err)
		}
		if !info.IsDir() {
			out = appendUnique(out, seen, top)
			continue
		}
		err = filepath.Walk(abs, func(path string, _ os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return fmt.Errorf("pack: rel path for %s: %w", path, err)
			}
			// tar wants forward slashes on every platform.
			rel = filepath.ToSlash(rel)
			out = appendUnique(out, seen, rel)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}

func appendUnique(out []string, seen map[string]struct{}, rel string) []string {
	if _, ok := seen[rel]; ok {
		return out
	}
	seen[rel] = struct{}{}
	return append(out, rel)
}

func writeEntry(tw *tar.Writer, archiveRoot, skillDir, rel string, ts time.Time) error {
	abs := filepath.Join(skillDir, filepath.FromSlash(rel))
	info, err := os.Lstat(abs)
	if err != nil {
		return fmt.Errorf("pack: stat %s: %w", abs, err)
	}
	// Symlinks are refused: they cross the archive boundary in ways
	// consumers cannot audit without extra care. Skills should not need
	// them at this stage.
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("pack: symlink not allowed in skill package: %s", rel)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("pack: build header for %s: %w", rel, err)
	}
	// Rewrite the archive path so consumers see <skill-name>/... instead
	// of skills/<skill-name>/...
	header.Name = archiveRoot + "/" + rel
	if info.IsDir() && !strings.HasSuffix(header.Name, "/") {
		header.Name += "/"
	}
	// Reproducibility: zero out everything the walk could leak.
	header.ModTime = ts
	header.AccessTime = time.Time{}
	header.ChangeTime = time.Time{}
	header.Uid = 0
	header.Gid = 0
	header.Uname = ""
	header.Gname = ""

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("pack: write header for %s: %w", rel, err)
	}
	if info.IsDir() {
		return nil
	}
	src, err := os.Open(abs) //nolint:gosec // path built from a vetted walk above.
	if err != nil {
		return fmt.Errorf("pack: open %s: %w", abs, err)
	}
	defer func() { _ = src.Close() }()
	if _, err := io.Copy(tw, src); err != nil {
		return fmt.Errorf("pack: copy %s: %w", rel, err)
	}
	return nil
}
