package skill_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SergioLacerda/skill-for-hire/internal/skill"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func manifestYAML(name, version string) string {
	return "apiVersion: skillsforhire.dev/v1\n" +
		"kind: Skill\n" +
		"metadata:\n" +
		"  name: " + name + "\n" +
		"  version: \"" + version + "\"\n" +
		"  owner: tester\n" +
		"  lifecycle: experimental\n" +
		"spec:\n" +
		"  entrypoint: SKILL.md\n" +
		"  capabilities: {}\n"
}

func rosterYAML(entries map[string]string) string {
	var b strings.Builder
	b.WriteString("schema-version: 1\nskills:\n")
	for name, version := range entries {
		b.WriteString("  " + name + ":\n")
		b.WriteString("    version: \"" + version + "\"\n")
	}
	return b.String()
}

func TestCheckRosterAcceptsMatchingVersions(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "skills", "alpha", "skill.yaml"), manifestYAML("alpha", "0.2.0"))
	rosterPath := filepath.Join(root, "roster.yaml")
	writeFile(t, rosterPath, rosterYAML(map[string]string{"alpha": "0.2.0"}))

	problems, err := skill.CheckRoster(rosterPath, []string{filepath.Join(root, "skills", "alpha")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
}

func TestCheckRosterReportsVersionDrift(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "skills", "alpha", "skill.yaml"), manifestYAML("alpha", "0.2.0"))
	rosterPath := filepath.Join(root, "roster.yaml")
	writeFile(t, rosterPath, rosterYAML(map[string]string{"alpha": "0.1.0"}))

	problems, err := skill.CheckRoster(rosterPath, []string{filepath.Join(root, "skills", "alpha")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], "alpha") ||
		!strings.Contains(problems[0], "0.1.0") || !strings.Contains(problems[0], "0.2.0") {
		t.Fatalf("expected one drift problem naming alpha, 0.1.0 and 0.2.0, got %v", problems)
	}
}

func TestCheckRosterReportsSkillMissingFromRoster(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "skills", "alpha", "skill.yaml"), manifestYAML("alpha", "0.2.0"))
	rosterPath := filepath.Join(root, "roster.yaml")
	writeFile(t, rosterPath, rosterYAML(map[string]string{"beta": "0.1.0"}))

	problems, err := skill.CheckRoster(rosterPath, []string{filepath.Join(root, "skills", "alpha")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], "alpha") {
		t.Fatalf("expected one missing-from-roster problem for alpha, got %v", problems)
	}
}

func TestCheckRosterRejectsUnreadableRoster(t *testing.T) {
	root := t.TempDir()
	_, err := skill.CheckRoster(filepath.Join(root, "missing.yaml"), nil)
	if err == nil {
		t.Fatal("expected error for a missing roster file, got nil")
	}
}

func TestCheckRosterRepositoryIsConsistent(t *testing.T) {
	problems, err := skill.CheckRoster("../../roster/skills.yaml",
		[]string{"../../skills/atlas", "../../skills/treasure-chest"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(problems) != 0 {
		t.Fatalf("repository roster is out of sync with skill manifests: %v", problems)
	}
}
