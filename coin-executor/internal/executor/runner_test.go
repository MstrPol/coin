package executor

import (
	"testing"

	"coin.local/coin-executor/internal/manifest"
)

func TestEnforcePublishPolicy_denied(t *testing.T) {
	m := &manifest.Manifest{
		Branching: &manifest.Branching{
			Name:    "trunk-based",
			Version: "1.0.0",
			Branches: []manifest.BranchRule{
				{Name: "feature", Pattern: `^feature/(?P<jira>[A-Z][A-Z0-9]*-\d+)(?:-.+)?$`, Versioning: manifest.BranchVersioning{Template: "v{base}-{jira}-snapshot-{n}"}, Publish: false},
			},
		},
	}
	t.Setenv("COIN_PUBLISH_REQUEST", "true")
	t.Setenv("GIT_BRANCH", "feature/PROJ-101")

	if err := enforcePublishPolicy(t.TempDir(), m); err == nil {
		t.Fatal("expected publish policy error")
	}
}

func TestEnforcePublishPolicy_allowed(t *testing.T) {
	m := &manifest.Manifest{
		Branching: &manifest.Branching{
			Name:    "trunk-based",
			Version: "1.0.0",
			Branches: []manifest.BranchRule{
				{Name: "release", Pattern: `^release/(?P<jira>[A-Z][A-Z0-9]*-\d+)(?:-.+)?$`, Versioning: manifest.BranchVersioning{Template: "v{base}-{jira}-rc-{n}"}, Publish: true},
			},
		},
	}
	t.Setenv("COIN_PUBLISH_REQUEST", "true")
	t.Setenv("GIT_BRANCH", "release/PROJ-404")

	if err := enforcePublishPolicy(t.TempDir(), m); err != nil {
		t.Fatal(err)
	}
}

func TestShouldSkipPublish_legacyTag(t *testing.T) {
	t.Setenv("TAG_NAME", "")
	m := &manifest.Manifest{
		Pipeline: manifest.Pipeline{
			Stages: []manifest.Stage{{ID: "publish", Name: "Publish", When: "tag"}},
		},
	}
	skip, _ := shouldSkipPublish(t.TempDir(), m)
	if !skip {
		t.Fatal("expected legacy skip")
	}
}
