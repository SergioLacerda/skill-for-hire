package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSkill(t *testing.T, root, name, version string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	manifest := "apiVersion: skillsforhire.dev/v1\nkind: Skill\nmetadata:\n  name: " + name +
		"\n  version: \"" + version + "\"\n  owner: tester\n  lifecycle: experimental\n" +
		"spec:\n  entrypoint: SKILL.md\n  capabilities: {}\n"
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name+"\n"), 0o600); err != nil {
		t.Fatalf("write entrypoint: %v", err)
	}
	return dir
}

func runValidate(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newValidateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestValidateRosterFlagFailsOnVersionDrift(t *testing.T) {
	root := t.TempDir()
	dir := writeSkill(t, root, "alpha", "0.2.0")
	roster := filepath.Join(root, "roster.yaml")
	if err := os.WriteFile(roster, []byte("skills:\n  alpha:\n    version: \"0.1.0\"\n"), 0o600); err != nil {
		t.Fatalf("write roster: %v", err)
	}

	out, err := runValidate(t, "--roster", roster, dir)
	if err == nil {
		t.Fatalf("expected failure on roster drift, output:\n%s", out)
	}
	if !strings.Contains(out, "roster version") {
		t.Fatalf("expected drift in output, got:\n%s", out)
	}
}

func TestValidateRosterFlagPassesWhenConsistent(t *testing.T) {
	root := t.TempDir()
	dir := writeSkill(t, root, "alpha", "0.2.0")
	roster := filepath.Join(root, "roster.yaml")
	if err := os.WriteFile(roster, []byte("skills:\n  alpha:\n    version: \"0.2.0\"\n"), 0o600); err != nil {
		t.Fatalf("write roster: %v", err)
	}

	if out, err := runValidate(t, "--roster", roster, dir); err != nil {
		t.Fatalf("expected success, got %v\n%s", err, out)
	}
}

func TestValidateWithoutRosterFlagIgnoresRoster(t *testing.T) {
	root := t.TempDir()
	dir := writeSkill(t, root, "alpha", "0.2.0")
	if out, err := runValidate(t, dir); err != nil {
		t.Fatalf("expected success without --roster, got %v\n%s", err, out)
	}
}
