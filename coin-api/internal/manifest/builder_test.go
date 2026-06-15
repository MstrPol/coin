package manifest

import "testing"

func TestBuilderStableHash(t *testing.T) {
	b := Builder{}
	release := sampleRelease()
	doc1, hash1, err := b.Build(release)
	if err != nil {
		t.Fatal(err)
	}
	doc2, hash2, err := b.Build(release)
	if err != nil {
		t.Fatal(err)
	}
	if hash1 != hash2 {
		t.Fatalf("hash mismatch: %s vs %s", hash1, hash2)
	}
	if doc1["manifestHash"] != doc2["manifestHash"] {
		t.Fatal("manifestHash field mismatch")
	}
}

func TestBuilderContentRefsURLShaped(t *testing.T) {
	b := Builder{}
	doc, _, err := b.Build(sampleRelease())
	if err != nil {
		t.Fatal(err)
	}

	schema, ok := doc["validateSchema"].(map[string]string)
	if !ok {
		t.Fatal("validateSchema missing")
	}
	if schema["url"] == "" || schema["sha256"] == "" {
		t.Fatalf("validateSchema not url-shaped: %#v", schema)
	}
	if _, hasGit := schema["gitRef"]; hasGit {
		t.Fatal("validateSchema must not contain gitRef")
	}

	pipeline, ok := doc["pipeline"].(map[string]any)
	if !ok {
		t.Fatal("pipeline missing")
	}
	rawStages, ok := pipeline["stages"].([]map[string]any)
	if !ok || len(rawStages) == 0 {
		t.Fatal("pipeline stages missing")
	}
	script, ok := rawStages[0]["script"].(map[string]string)
	if !ok || script["url"] == "" || script["sha256"] == "" {
		t.Fatalf("stage script not url-shaped: %#v", rawStages[0])
	}
	if _, hasOrch := doc["orchestration"]; hasOrch {
		t.Fatal("manifest must not contain orchestration.bundle")
	}
	jnlp, ok := doc["jnlp"].(map[string]string)
	if !ok || jnlp["image"] == "" {
		t.Fatalf("jnlp missing: %#v", doc["jnlp"])
	}
}

func sampleRelease() GPRelease {
	return GPRelease{
		Name:    "go-app",
		Version: "1.0.0",
		Parts: Composition{
			ExecutorVersion:  "0.1.0",
			AgentImage:       "localhost:8082/coin-docker/ci-go:1.22.5",
			AgentDigest:      "sha256:deadbeef",
			JnlpImage:        "jenkins/inbound-agent:latest",
			JnlpDigest:       "sha256:jnlp",
			ExecutorURL:      "http://localhost:8081/repository/maven-releases/coin/executor/coin-executor/0.1.0/coin-executor-0.1.0-linux-amd64.bin",
			ExecutorSHA256:   "sha256:abc",
			LibName:          "coin-lib",
			LibVersion:       "1.0.0",
			GPContentName:    "go-app",
			GPContentVersion: "1.0.0",
		},
		Content: ContentBundle{
			SchemaArtifactKey:     "schemas/config.v2.schema.json",
			SchemaSHA256:          "sha256:schema",
			DockerfileArtifactKey: "dockerfiles/go-runtime.Dockerfile",
			DockerfileSHA256:      "sha256:dockerfile",
			Stages: []StageScript{
				{Name: "validate", ArtifactKey: "scripts/validate.sh", SHA256: "sha256:v"},
				{Name: "test", ArtifactKey: "scripts/test.sh", SHA256: "sha256:t"},
				{Name: "build", ArtifactKey: "scripts/build.sh", SHA256: "sha256:b"},
				{Name: "publish", When: "tag", ArtifactKey: "scripts/publish.sh", SHA256: "sha256:p"},
			},
		},
	}
}
