package componentpackage

import (
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

func TestValidateDraftPackage_gpContent(t *testing.T) {
	t.Parallel()
	validYAML := `schemaVersion: 2
name: go-app
kind: gp-content
capabilities:
  deliverables:
    - image
    - artifact
build:
  engine: buildkit
  buildkit:
    targets:
      validate: validate
      test: test
      image: runtime
      artifact: artifact
    cacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:buildkit"
pipeline:
  stages:
    - id: test
      name: Test
artifacts:
  validateSchema: schemas/config.v2.schema.json
  containerfile: dockerfiles/Containerfile
`
	res := ValidateDraftPackage("gp-content", "go-app", "1.0.0", []DraftArtifact{
		{Path: "content.yaml", Body: []byte(validYAML), SHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{Path: "dockerfiles/Containerfile", Body: []byte("FROM scratch\n"), SHA256: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
	})
	if !res.Valid {
		t.Fatalf("expected valid, issues: %+v", res.Issues)
	}

	res = ValidateDraftPackage("gp-content", "go-app", "1.0.0", []DraftArtifact{
		{Path: "content.yaml", Body: []byte(validYAML), SHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	})
	if res.Valid {
		t.Fatal("expected missing containerfile")
	}

	v1YAML := `name: go-app
kind: gp-content
build:
  engine: buildkit
`
	res = ValidateDraftPackage("gp-content", "go-app", "1.0.0", []DraftArtifact{
		{Path: "content.yaml", Body: []byte(v1YAML), SHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	})
	if res.Valid {
		t.Fatal("expected v1 schema rejection")
	}
}
