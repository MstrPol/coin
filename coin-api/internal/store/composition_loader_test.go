package store

import (
	"encoding/json"
	"testing"

	"coin.local/coin-api/internal/manifest"
)

func TestContentBundleFromV2Manifest(t *testing.T) {
	raw := json.RawMessage(`{
		"schemaVersion": 2,
		"package": {"url": "http://nexus/pkg.zip", "sha256": "abc"},
		"manifest": {
			"build": {"engine": "buildkit", "buildkit": {"dockerfile": "dockerfiles/Containerfile", "cacheRefTemplate": "{{registryHost}}/cache"}},
			"pipeline": {"stages": [{"id": "test", "name": "Test"}]},
			"validateSchema": {"artifactKey": "schemas/config.v2.schema.json", "sha256": "s1"},
			"containerfile": {"artifactKey": "dockerfiles/Containerfile", "sha256": "s2"}
		}
	}`)
	meta := gpContentMetadata{URL: "http://legacy", SHA256: "legacy"}
	bundle, err := contentBundleFromV2Manifest(meta, raw)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.BuildEngine != "buildkit" {
		t.Fatalf("engine=%q", bundle.BuildEngine)
	}
	if len(bundle.Stages) != 1 || bundle.Stages[0].ID != "test" {
		t.Fatalf("stages=%+v", bundle.Stages)
	}
	if bundle.SchemaArtifactKey != "schemas/config.v2.schema.json" {
		t.Fatalf("schema=%q", bundle.SchemaArtifactKey)
	}
}

func TestApplyCompositionSlotRegistry(t *testing.T) {
	var parts manifest.Composition
	applyCompositionSlot(&parts, "agent", "coin-agent", "1.0.0", map[string]any{
		"image": "nexus:8082/coin-docker/coin-agent:1.0.0", "digest": "sha256:abc",
	})
	applyCompositionSlot(&parts, "gp-content", "go-app", "1.0.0", nil)
	if parts.AgentImage == "" || parts.GPContentName != "go-app" {
		t.Fatalf("parts=%+v", parts)
	}
}
