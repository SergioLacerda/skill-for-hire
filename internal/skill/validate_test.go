package skill_test

import (
	"testing"

	"github.com/SergioLacerda/skill-for-hire/internal/skill"
)

func TestValidateAtlasHelloWorld(t *testing.T) {
	report, err := skill.Validate("../../skills/atlas")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.OK() {
		t.Fatalf("expected atlas hello-world to pass validation, got errors: %v", report.Errors)
	}
	if report.Manifest.Metadata.Name != "atlas" {
		t.Fatalf("expected name atlas, got %q", report.Manifest.Metadata.Name)
	}
}

func TestValidateTreasureChestHelloWorld(t *testing.T) {
	report, err := skill.Validate("../../skills/treasure-chest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.OK() {
		t.Fatalf("expected treasure-chest hello-world to pass validation, got errors: %v", report.Errors)
	}
}

func TestValidateRejectsMissingDir(t *testing.T) {
	_, err := skill.Validate("../../skills/does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing skill directory, got nil")
	}
}
