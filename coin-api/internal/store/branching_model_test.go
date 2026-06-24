package store

import (
	"encoding/json"
	"testing"

	"coin.local/coin-api/internal/manifest"
)

func TestBranchingRulesFromV2Manifest(t *testing.T) {
	raw := json.RawMessage(`{
		"schemaVersion": 2,
		"package": {"url": "http://nexus/pkg.zip", "sha256": "abc"},
		"manifest": {
			"branching": {
				"name": "trunk-based",
				"trunk": {"branch": "main"},
				"branchTypes": ["feature", "release"],
				"publish": {"when": "tag"}
			}
		}
	}`)
	rules, err := branchingRulesFromV2Manifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if rules["name"] != "trunk-based" {
		t.Fatalf("name=%v", rules["name"])
	}
	trunk, ok := rules["trunk"].(map[string]any)
	if !ok || trunk["branch"] != "main" {
		t.Fatalf("trunk=%v", rules["trunk"])
	}
}

func TestDefaultBranchingModelForGP(t *testing.T) {
	if DefaultBranchingModelForGP("go-app") != "trunk-based" {
		t.Fatal("go-app expected trunk-based")
	}
	if DefaultBranchingModelForGP("go-lib") != "semver-tag" {
		t.Fatal("go-lib expected semver-tag")
	}
}

func TestCanonicalGPSlotsFive(t *testing.T) {
	slots := CanonicalGPSlots("go-app")
	if len(slots) != 5 || slots[4].Name != "trunk-based" {
		t.Fatalf("slots=%+v", slots)
	}
}

func TestApplyBranchingComposition(t *testing.T) {
	var parts manifest.Composition
	applyCompositionSlot(&parts, "branching-model", "trunk-based", "1.0.0", nil)
	if parts.BranchingModelName != "trunk-based" || parts.BranchingModelVersion != "1.0.0" {
		t.Fatalf("parts=%+v", parts)
	}
}
