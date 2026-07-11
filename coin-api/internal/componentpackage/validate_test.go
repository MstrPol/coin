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
	full := json.RawMessage(`{"schemaVersion":2,"package":{"url":"http://x","sha256":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`)
	if !IsContentRefV2(full) {
		t.Fatal("expected full v2")
	}
	pgOnly := json.RawMessage(`{"schemaVersion":2,"manifest":{"branching":{"model":"trunk-based"}}}`)
	if IsContentRefV2(pgOnly) {
		t.Fatal("PG-only should not be full v2")
	}
	if !IsContentRefV2Envelope(pgOnly) {
		t.Fatal("expected v2 envelope")
	}
	if !IsRegisteredForCanary(pgOnly) {
		t.Fatal("expected registered for canary")
	}
	if IsContentRefV2(json.RawMessage(`{"artifactKey":"x"}`)) {
		t.Fatal("legacy should not be v2")
	}
}

func TestBuildContentRefV2PGOnly(t *testing.T) {
	t.Parallel()
	raw, err := BuildContentRefV2PGOnly(map[string]any{"branching": map[string]any{"model": "trunk-based"}})
	if err != nil {
		t.Fatal(err)
	}
	if HasPackageURL(raw) {
		t.Fatal("PG-only should not have package")
	}
	if err := ValidateContentRefOnWrite(raw); err != nil {
		t.Fatalf("PG-only should validate on write: %v", err)
	}
}
