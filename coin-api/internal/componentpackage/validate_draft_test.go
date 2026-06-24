package componentpackage

import (
	"testing"
)

func TestValidateDraftPackage_branchingModel(t *testing.T) {
	t.Parallel()
	validYAML := `schemaVersion: 1
name: trunk-based
trunk:
  branch: main
branchTypes:
  - feature
  - bugfix
  - release
versioning:
  tagPrefix: v
  qualifiers:
    snapshot:
      enabled: true
    rc:
      enabled: true
      releaseBranchesOnly: true
publish:
  when: tag
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
}

func TestValidateDraftPackage_gpContent(t *testing.T) {
	t.Parallel()
	validYAML := `name: go-app
kind: gp-content
build:
  engine: buildkit
pipeline:
  stages:
    - id: test
      name: Test
validateSchema:
  artifactKey: schemas/config.v2.schema.json
containerfile:
  artifactKey: dockerfiles/Containerfile
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
}
