package componentpackage

import (
	"strings"
	"testing"
)

func TestValidateDraftPackage_branchingModel(t *testing.T) {
	t.Parallel()
	validYAML := `schemaVersion: 2
name: trunk-based
branches:
  - name: main
    pattern: ^main$|^master$
    versioning:
      template: "v{base}-main-snapshot-{n}"
    publish: false
  - name: release
    pattern: ^release/(?P<jira>[A-Z][A-Z0-9]*-\d+)(?:-.+)?$
    versioning:
      template: "v{base}-{jira}-rc-{n}"
    publish: true
`
	res := ValidateDraftPackage("branching-model", "trunk-based", "1.0.0", []DraftArtifact{
		{Path: "model.yaml", Body: []byte(validYAML), SHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	})
	if !res.Valid {
		t.Fatalf("expected valid, issues: %+v", res.Issues)
	}

	res = ValidateDraftPackage("branching-model", "trunk-based", "1.0.0", nil)
	if res.Valid || len(res.Issues) == 0 {
		t.Fatal("expected missing artifacts error")
	}

	res = ValidateDraftPackage("branching-model", "trunk-based", "1.0.0", []DraftArtifact{
		{Path: "model.yaml", Body: []byte("name: x\n"), SHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	})
	if res.Valid {
		t.Fatal("expected invalid yaml/model")
	}

	v1YAML := `schemaVersion: 1
name: trunk-based
trunk:
  branch: main
`
	res = ValidateDraftPackage("branching-model", "trunk-based", "1.0.0", []DraftArtifact{
		{Path: "model.yaml", Body: []byte(v1YAML), SHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	})
	if res.Valid {
		t.Fatal("expected v1 schema rejection")
	}
}

func TestValidateDraftPackage_gpContentDeprecated(t *testing.T) {
	t.Parallel()
	res := ValidateDraftPackage("gp-content", "go-app", "1.0.0", []DraftArtifact{
		{Path: "content.yaml", Body: []byte("schemaVersion: 3\nname: go-app\n"), SHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	})
	if res.Valid {
		t.Fatal("expected gp-content type rejection")
	}
	found := false
	for _, iss := range res.Issues {
		if iss.Field == "componentType" && strings.Contains(iss.Message, "removed") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected deprecation issue, got %+v", res.Issues)
	}
}
