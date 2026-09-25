package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCommand executes the CLI with the given args against a fresh cobra
// tree, capturing stdout. Returns stdout, stderr and the RunE error.
func runCommand(args ...string) (string, string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := newRootCmd()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func TestVersionCommandPrintsVersion(t *testing.T) {
	out, _, err := runCommand("version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("version output is empty")
	}
}

func TestStatusHumanFormatIncludesProviderAndCapabilities(t *testing.T) {
	tmp := t.TempDir()
	out, _, err := runCommand("status", "--root", tmp)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	for _, want := range []string{
		"provider:",
		"treasure-chest",
		"capabilities:",
		"knowledge.mine",
		"knowledge.search",
		"schema:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("status output missing %q\nfull:\n%s", want, out)
		}
	}
}

func TestStatusJSONMatchesEnvelopeContract(t *testing.T) {
	tmp := t.TempDir()
	out, _, err := runCommand("status", "--root", tmp, "--json")
	if err != nil {
		t.Fatalf("status --json: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("status --json is not valid JSON: %v\n%s", err, out)
	}
	for _, key := range []string{"Provider", "Version", "SchemaVersion", "Capabilities", "Healthy"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("status JSON missing key %q", key)
		}
	}
	if provider, _ := payload["Provider"].(string); provider != "treasure-chest" {
		t.Errorf("provider = %v, want treasure-chest", payload["Provider"])
	}
}

func TestPrepareFailsWithoutRoot(t *testing.T) {
	// --root is optional at the flag layer (defaults to "."), but if a
	// caller passes an empty explicit root, Prepare must reject it.
	_, _, err := runCommand("prepare", "--root", "")
	if err == nil {
		t.Fatal("expected prepare with empty --root to fail")
	}
}

func TestSearchEmptyWorkspaceReturnsZeroItems(t *testing.T) {
	tmp := t.TempDir()
	out, _, err := runCommand("search", "--root", tmp, "--intent", "anything")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !strings.Contains(out, "0 item(s) matched") {
		t.Fatalf("search on empty workspace should report zero matches; got:\n%s", out)
	}
}

func TestExplainUnknownIDReportsDiscarded(t *testing.T) {
	tmp := t.TempDir()
	out, _, err := runCommand("explain", "no-such-jewel", "--root", tmp)
	if err != nil {
		t.Fatalf("explain: %v", err)
	}
	if !strings.Contains(out, "decision: discarded") {
		t.Fatalf("explain of unknown id should be discarded; got:\n%s", out)
	}
}

func TestPrepareAgainstMinimalChestWorkspaceIndexes(t *testing.T) {
	// Prepare only reads governed.yaml + active.yaml + jewels/potions;
	// a bare workspace with an empty governed file must round-trip
	// without error and count 0 items.
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, ".strategist", "chest", "governed.yaml"), "governed_chests: []\n")
	writeFile(t, filepath.Join(tmp, ".strategist", "chest", "active.yaml"), "active_chests: []\n")
	writeFile(t, filepath.Join(tmp, ".strategist", "chest", "indexed.yaml"), "indexed: []\n")

	out, _, err := runCommand("prepare", "--root", tmp)
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if !strings.Contains(out, "prepared root=") {
		t.Fatalf("prepare output missing header; got:\n%s", out)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
