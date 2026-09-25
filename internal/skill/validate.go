package skill

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// nameRegex is the kebab-case identifier rule shared by ORKA and the Skills
// for Hire schema: letters/digits, non-consecutive single hyphens, no
// leading or trailing hyphen.
var nameRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// semverRegex is a pragmatic SemVer check — enough to catch typos like
// "1.0" or "1.a.0" without pulling a full parser.
var semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)

var allowedLifecycles = map[string]struct{}{
	"experimental": {},
	"beta":         {},
	"stable":       {},
	"deprecated":   {},
	"retired":      {},
}

var allowedCapabilityModes = map[string]struct{}{
	"denied":          {},
	"read-only":       {},
	"workspace-read":  {},
	"workspace-write": {},
	"allowlist":       {},
}

// Report bundles findings so the CLI can print them all at once instead of
// stopping on the first one.
type Report struct {
	SkillDir string
	Manifest *Manifest
	Errors   []string
	Warnings []string
}

func (r *Report) addErr(format string, args ...any) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

func (r *Report) addWarn(format string, args ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

// OK reports whether the skill passes package-static validation.
func (r *Report) OK() bool { return len(r.Errors) == 0 }

// Validate performs package-static conformance checks against a skill
// directory. It never runs the skill, never makes network calls, and
// treats missing evidence as an error, not a silent pass.
func Validate(skillDir string) (*Report, error) {
	report := &Report{SkillDir: skillDir}

	info, err := os.Stat(skillDir)
	if err != nil {
		return nil, fmt.Errorf("stat skill dir: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("skill path is not a directory: %s", skillDir)
	}

	folderName := filepath.Base(filepath.Clean(skillDir))
	if !nameRegex.MatchString(folderName) {
		report.addErr("folder name %q is not kebab-case (^[a-z0-9]+(-[a-z0-9]+)*$)", folderName)
	}

	manifestPath := filepath.Join(skillDir, "skill.yaml")
	manifest, err := Load(manifestPath)
	if err != nil {
		report.addErr("%v", err)
		return report, nil
	}
	report.Manifest = manifest

	validateHeader(report, manifest)
	validateMetadata(report, manifest, folderName)
	validateSpec(report, skillDir, manifest)

	return report, nil
}

func validateHeader(r *Report, m *Manifest) {
	if m.APIVersion != APIVersion {
		r.addErr("apiVersion must be %q, got %q", APIVersion, m.APIVersion)
	}
	if m.Kind != Kind {
		r.addErr("kind must be %q, got %q", Kind, m.Kind)
	}
}

func validateMetadata(r *Report, m *Manifest, folderName string) {
	if m.Metadata.Name == "" {
		r.addErr("metadata.name is required")
	} else if !nameRegex.MatchString(m.Metadata.Name) {
		r.addErr("metadata.name %q is not kebab-case", m.Metadata.Name)
	} else if m.Metadata.Name != folderName {
		r.addErr("metadata.name %q must match folder name %q", m.Metadata.Name, folderName)
	}

	if m.Metadata.Version == "" {
		r.addErr("metadata.version is required")
	} else if !semverRegex.MatchString(m.Metadata.Version) {
		r.addErr("metadata.version %q is not valid SemVer", m.Metadata.Version)
	}

	if m.Metadata.Owner == "" {
		r.addErr("metadata.owner is required")
	}
	if m.Metadata.Lifecycle == "" {
		r.addErr("metadata.lifecycle is required")
	} else if _, ok := allowedLifecycles[m.Metadata.Lifecycle]; !ok {
		r.addErr("metadata.lifecycle %q is not one of experimental|beta|stable|deprecated|retired", m.Metadata.Lifecycle)
	}
}

func validateSpec(r *Report, skillDir string, m *Manifest) {
	if m.Spec.Entrypoint == "" {
		r.addErr("spec.entrypoint is required")
	} else {
		entryPath := filepath.Join(skillDir, m.Spec.Entrypoint)
		if _, err := os.Stat(entryPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				r.addErr("spec.entrypoint %q does not exist relative to skill root", m.Spec.Entrypoint)
			} else {
				r.addErr("stat entrypoint %s: %v", m.Spec.Entrypoint, err)
			}
		}
	}

	// deny-by-default only makes sense when the capabilities block is
	// present at all; an empty map means "everything denied" and is fine.
	if m.Spec.Capabilities == nil {
		r.addWarn("spec.capabilities is missing; interpreting as deny-by-default")
	} else {
		validateCapabilities(r, m.Spec.Capabilities)
	}

	if m.Spec.Composition.Provides == nil {
		r.addWarn("spec.composition.provides is empty; consumers cannot select this skill by capability")
	}
}

func validateCapabilities(r *Report, caps map[string]any) {
	for key, raw := range caps {
		block, ok := raw.(map[string]any)
		if !ok {
			// destructiveActions.confirmation is a plain map with no mode;
			// skip anything that isn't a capability object.
			continue
		}
		mode, ok := block["mode"].(string)
		if !ok {
			// Some capability blocks (e.g. destructiveActions) don't
			// require a mode — only check when one is present.
			continue
		}
		if _, ok := allowedCapabilityModes[mode]; !ok {
			r.addErr("spec.capabilities.%s.mode %q is not one of denied|read-only|workspace-read|workspace-write|allowlist", key, mode)
		}
	}
}
