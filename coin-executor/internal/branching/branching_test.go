package branching

import (
	"testing"

	"coin.local/coin-executor/internal/manifest"
)

func trunkBasedModel() *Model {
	return FromManifest(&manifest.Manifest{
		Branching: &manifest.Branching{
			Name:    "trunk-based",
			Version: "1.0.0",
			Trunk:   manifest.BranchingTrunk{Branch: "main"},
			BranchTypes: []string{"feature", "bugfix", "release"},
			Versioning: manifest.BranchingVersion{
				TagPrefix: "v",
				Qualifiers: manifest.BranchingQualifiers{
					Snapshot: manifest.BranchingQualifierToggle{Enabled: true},
					RC: manifest.BranchingRCQualifier{
						Enabled: true, ReleaseBranchesOnly: true,
					},
				},
			},
			Publish: manifest.BranchingPublish{When: "tag"},
		},
	})
}

func semverTagModel() *Model {
	return FromManifest(&manifest.Manifest{
		Branching: &manifest.Branching{
			Name:    "semver-tag",
			Version: "1.0.0",
			Trunk:   manifest.BranchingTrunk{Branch: "main"},
			BranchTypes: []string{"feature", "bugfix", "release"},
			Versioning: manifest.BranchingVersion{
				TagPrefix: "v",
				Qualifiers: manifest.BranchingQualifiers{
					Snapshot: manifest.BranchingQualifierToggle{Enabled: false},
					RC:       manifest.BranchingRCQualifier{Enabled: false},
				},
			},
			Publish: manifest.BranchingPublish{When: "tag"},
		},
	})
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

func TestShouldPublish_trunkBased(t *testing.T) {
	t.Parallel()
	m := trunkBasedModel()

	ok, _ := ShouldPublish(m, GitContext{Branch: "feature/PROJ-101"})
	if ok {
		t.Fatal("feature without tag must not publish")
	}

	ok, _ = ShouldPublish(m, GitContext{
		Branch: "release/PROJ-404", TagName: "v1.0.0-PROJ-404-rc-1",
	})
	if !ok {
		t.Fatal("rc tag on release branch must publish")
	}

	ok, _ = ShouldPublish(m, GitContext{
		Branch: "release/PROJ-404", TagName: "v1.0.0-PROJ-404-snapshot-1",
	})
	if ok {
		t.Fatal("snapshot tag must not publish with trunk-based tag policy")
	}
}

func TestShouldPublish_semverTag(t *testing.T) {
	t.Parallel()
	m := semverTagModel()

	ok, _ := ShouldPublish(m, GitContext{Branch: "main", TagName: "v1.2.3"})
	if !ok {
		t.Fatal("semver tag must publish")
	}

	ok, _ = ShouldPublish(m, GitContext{Branch: "main", TagName: "v1.2.3-rc-1"})
	if ok {
		t.Fatal("non-semver tag must not publish for semver-tag model")
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

func TestResolveVersion_semverTag(t *testing.T) {
	t.Parallel()
	m := semverTagModel()
	g := GitContext{Branch: "main"}
	v, err := ResolveVersion(m, g)
	if err != nil {
		t.Fatal(err)
	}
	if v != defaultVersion {
		t.Fatalf("expected %s, got %q", defaultVersion, v)
	}
}

func TestBump_rcRequiresReleaseBranch(t *testing.T) {
	t.Parallel()
	m := trunkBasedModel()
	_, err := Bump(m, GitContext{Branch: "feature/PROJ-101"}, "rc")
	if err == nil {
		t.Fatal("expected rc bump error on feature branch")
	}
}
