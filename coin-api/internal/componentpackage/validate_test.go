package componentpackage

import (
	"encoding/json"
	"testing"
)

func TestValidateContentRefV2(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{
		"schemaVersion": 2,
		"package": {
			"url": "http://nexus:8081/repository/maven-releases/coin/gp-content/go-app/1.0.0/package.manifest.json",
			"sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
		"manifest": {"build": {"engine": "buildkit"}}
	}`)
	ref, err := ValidateContentRefV2(raw)
	if err != nil {
		t.Fatal(err)
	}
	if ref.Package.URL == "" || ref.Manifest == nil {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}

func TestValidatePackageManifest(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{
		"schemaVersion": 1,
		"componentType": "branching-model",
		"componentName": "trunk-based",
		"version": "1.0.0",
		"files": [{"path": "model.yaml", "sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "role": "primary"}]
	}`)
	m, err := ValidatePackageManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m.ComponentType != "branching-model" || len(m.Files) != 1 {
		t.Fatalf("unexpected manifest: %+v", m)
	}
}

func TestValidateContentRefOnWrite(t *testing.T) {
	t.Parallel()
	if err := ValidateContentRefOnWrite(json.RawMessage(`{"artifactKey":"schema/x.json"}`)); err != nil {
		t.Fatalf("legacy should pass: %v", err)
	}
	if err := ValidateContentRefOnWrite(nil); err != nil {
		t.Fatalf("empty should pass: %v", err)
	}
	bad := json.RawMessage(`{"schemaVersion":2,"package":{"url":"not-a-url","sha256":"bad"}}`)
	if err := ValidateContentRefOnWrite(bad); err == nil {
		t.Fatal("expected v2 validation error")
	}
}

func TestIsContentRefV2(t *testing.T) {
	t.Parallel()
	if !IsContentRefV2(json.RawMessage(`{"schemaVersion":2,"package":{"url":"http://x","sha256":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`)) {
		t.Fatal("expected v2")
	}
	if IsContentRefV2(json.RawMessage(`{"artifactKey":"x"}`)) {
		t.Fatal("legacy should not be v2")
	}
}
