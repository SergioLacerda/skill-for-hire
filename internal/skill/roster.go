package skill

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// rosterFile is the subset of roster/skills.yaml the cross-check reads:
// each catalog entry's declared version, keyed by skill name.
type rosterFile struct {
	Skills map[string]struct {
		Version string `yaml:"version"`
	} `yaml:"skills"`
}

// CheckRoster compares the catalog at rosterPath with the manifests of the
// given skill directories and returns one problem per disagreement: a skill
// absent from the roster, or a roster version that differs from the
// manifest's metadata.version. An unreadable roster or manifest is an error,
// not a problem, so a broken catalog never reads as "consistent".
func CheckRoster(rosterPath string, skillDirs []string) ([]string, error) {
	data, err := os.ReadFile(rosterPath) //nolint:gosec // path comes from a CLI arg the operator provided.
	if err != nil {
		return nil, fmt.Errorf("read roster %s: %w", rosterPath, err)
	}
	var roster rosterFile
	if err := yaml.Unmarshal(data, &roster); err != nil {
		return nil, fmt.Errorf("parse roster %s: %w", rosterPath, err)
	}

	var problems []string
	for _, dir := range skillDirs {
		manifest, err := Load(dir + "/skill.yaml")
		if err != nil {
			return nil, err
		}
		name := manifest.Metadata.Name
		entry, ok := roster.Skills[name]
		switch {
		case !ok:
			problems = append(problems, fmt.Sprintf("skill %q (%s) is not listed in %s", name, dir, rosterPath))
		case entry.Version != manifest.Metadata.Version:
			problems = append(problems, fmt.Sprintf(
				"skill %q: roster version %q does not match manifest version %q (%s)",
				name, entry.Version, manifest.Metadata.Version, dir))
		}
	}
	sort.Strings(problems)
	return problems, nil
}
