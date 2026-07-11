package branching

import (
	"testing"

	"coin.local/coin-executor/internal/manifest"
)

func trunkBasedManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Branching: &manifest.Branching{
			Name:    "trunk-based",
			Version: "1.0.0",
			Branches: []manifest.BranchRule{
				{Name: "main", Pattern: `^main$|^master$`, Versioning: manifest.BranchVersioning{Template: "v{base}-main-snapshot-{n}"}, Publish: false},
				{Name: "feature", Pattern: `^feature/(?P<jira>[A-Z][A-Z0-9]*-\d+)(?:-.+)?$`, Versioning: manifest.BranchVersioning{Template: "v{base}-{jira}-snapshot-{n}"}, Publish: false},
				{Name: "release", Pattern: `^release/(?P<jira>[A-Z][A-Z0-9]*-\d+)(?:-.+)?$`, Versioning: manifest.BranchVersioning{Template: "v{base}-{jira}-rc-{n}"}, Publish: true},
			},
		},
	}
}

func trunkBasedModel() *Model {
	return FromManifest(trunkBasedManifest())
}

func TestValidateBranch_trunkBased(t *testing.T) {
	t.Parallel()
	m := trunkBasedModel()
	for _, branch := range []string{"main", "feature/PROJ-101", "release/PROJ-404"} {
		if err := ValidateBranch(m, branch); err != nil {
			t.Fatalf("branch %q: %v", branch, err)
		}
	}
	if err := ValidateBranch(m, "hotfix/PROJ-1"); err == nil {
		t.Fatal("expected hotfix to fail")
	}
}

func TestCheckPublishAllowed_deniedOnFeature(t *testing.T) {
	m := trunkBasedModel()
	t.Setenv("COIN_PUBLISH_REQUEST", "true")
	err := CheckPublishAllowed(m, GitContext{Branch: "feature/PROJ-101"})
	if err == nil {
		t.Fatal("expected publish denied on feature")
	}
}

func TestCheckPublishAllowed_allowedOnRelease(t *testing.T) {
	m := trunkBasedModel()
	t.Setenv("COIN_PUBLISH_REQUEST", "true")
	if err := CheckPublishAllowed(m, GitContext{Branch: "release/PROJ-404"}); err != nil {
		t.Fatal(err)
	}
}

func TestResolveVersion_trunkBasedFromTag(t *testing.T) {
	t.Parallel()
	m := trunkBasedModel()
	v, err := ResolveVersion(m, GitContext{
		Branch: "release/PROJ-404", TagName: "v1.5.0-PROJ-404-rc-2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if v != "1.5.0-PROJ-404-rc-2" {
		t.Fatalf("got %q", v)
	}
}

func TestMatch_firstWins(t *testing.T) {
	t.Parallel()
	m := trunkBasedModel()
	match, err := m.Match("feature/PROJ-101")
	if err != nil {
		t.Fatal(err)
	}
	if match.Rule.Name != "feature" {
		t.Fatalf("got rule %q", match.Rule.Name)
	}
}

func TestPreviewScenarios_publishDenied(t *testing.T) {
	t.Parallel()
	m := trunkBasedModel()
	res, err := PreviewScenarios(m, []PreviewScenario{{
		ID: "f1", Branch: "feature/PROJ-101", RequestPublish: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 1 || res.Results[0].PublishOutcome != string(PublishDenied) {
		t.Fatalf("%+v", res.Results)
	}
}

func TestGitFromEnv_prefersChangeBranch(t *testing.T) {
	t.Setenv("CHANGE_BRANCH", "feature/PROJ-101")
	t.Setenv("BRANCH_NAME", "PR-42")
	g, err := GitFromEnv("")
	if err != nil {
		t.Fatal(err)
	}
	if g.Branch != "feature/PROJ-101" {
		t.Fatalf("got %q", g.Branch)
	}
}
