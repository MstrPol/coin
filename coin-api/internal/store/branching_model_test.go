package store

import (
	"encoding/json"
	"testing"
)

func TestBranchingRulesFromRawRefPGOnly(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{
		"schemaVersion": 2,
		"manifest": {
			"branching": {
				"model": "trunk-based",
				"versioning": "semver-tag"
			}
		}
	}`)
	rules, err := branchingRulesFromRawRef(raw)
	if err != nil {
		t.Fatal(err)
	}
	if rules["model"] != "trunk-based" {
		t.Fatalf("unexpected rules: %+v", rules)
	}
}

func TestBranchingRulesFromRawRefPublished(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{
		"schemaVersion": 2,
		"package": {
			"url": "http://nexus/package.manifest.json",
			"sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
		"manifest": {
			"branching": {"model": "semver-tag"}
		}
	}`)
	rules, err := branchingRulesFromRawRef(raw)
	if err != nil {
		t.Fatal(err)
	}
	if rules["model"] != "semver-tag" {
		t.Fatalf("unexpected rules: %+v", rules)
	}
}
