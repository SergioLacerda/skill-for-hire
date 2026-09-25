// Package skill parses and validates a skill.yaml manifest against the
// skillsforhire.dev/v1 schema.
package skill

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// APIVersion is the only accepted apiVersion for now.
const APIVersion = "skillsforhire.dev/v1"

// Kind is the only accepted kind for now.
const Kind = "Skill"

// Manifest mirrors doctrine/skill.schema.json. Fields cover only what the
// current CLI needs; unknown fields are preserved but ignored.
type Manifest struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

// Metadata is the schema's metadata block: identity, version, ownership
// and lifecycle. Present in every valid Skill manifest.
type Metadata struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	DisplayName string   `yaml:"displayName,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Owner       string   `yaml:"owner"`
	Lifecycle   string   `yaml:"lifecycle"`
	Tags        []string `yaml:"tags,omitempty"`
}

// Spec is the operational contract: entrypoint, declared capabilities,
// composition and distribution details.
type Spec struct {
	Entrypoint   string         `yaml:"entrypoint"`
	Capabilities map[string]any `yaml:"capabilities"`
	Composition  Composition    `yaml:"composition,omitempty"`
	Dependencies Dependencies   `yaml:"dependencies,omitempty"`
	Distribution map[string]any `yaml:"distribution,omitempty"`
	Extra        map[string]any `yaml:",inline"`
}

// Composition describes the capabilities the skill provides and any it
// conflicts with — the surface consumers select against.
type Composition struct {
	Provides      []string `yaml:"provides,omitempty"`
	ConflictsWith []string `yaml:"conflictsWith,omitempty"`
	Priority      int      `yaml:"priority,omitempty"`
}

// Dependencies holds required and optional inter-skill dependencies.
type Dependencies struct {
	Required []Dependency `yaml:"required,omitempty"`
	Optional []Dependency `yaml:"optional,omitempty"`
}

// Dependency is one entry in the required or optional list: a skill
// name plus a SemVer range and, optionally, the capability it satisfies.
type Dependency struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Capability string `yaml:"capability,omitempty"`
}

// Load reads and decodes a skill.yaml manifest.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from a CLI arg the operator provided.
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", path, err)
	}
	return &m, nil
}
