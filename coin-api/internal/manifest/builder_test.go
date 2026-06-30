package manifest

import "testing"

func TestBuilderStableHash(t *testing.T) {
	b := Builder{}
	release := sampleRelease()
	opts := BuildOptions{Project: "demo-go-app", RegistryHost: "localhost:8082"}
	doc1, hash1, err := b.Build(release, opts)
	if err != nil {
		t.Fatal(err)
	}
	doc2, hash2, err := b.Build(release, opts)
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

func TestBuilderBuildEngineContract(t *testing.T) {
	b := Builder{}
	doc, _, err := b.Build(sampleRelease(), BuildOptions{Project: "demo-go-app", RegistryHost: "localhost:8082"})
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

	build, ok := doc["build"].(map[string]any)
	if !ok {
		t.Fatal("build missing")
	}
	if build["engine"] != "buildkit" {
		t.Fatalf("unexpected engine: %#v", build["engine"])
	}
	buildkit, ok := build["buildkit"].(map[string]any)
	if !ok {
		t.Fatal("build.buildkit missing")
	}
	if buildkit["dockerfile"] != ".coin/Containerfile" {
		t.Fatalf("unexpected dockerfile path: %#v", buildkit["dockerfile"])
	}
	cacheRef, _ := buildkit["cacheRef"].(string)
	if cacheRef != "localhost:8082/coin-cache/demo-go-app:buildkit" {
		t.Fatalf("unexpected cacheRef: %q", cacheRef)
	}
	cf, ok := buildkit["containerfile"].(map[string]string)
	if !ok || cf["url"] == "" {
		t.Fatalf("containerfile ref missing: %#v", buildkit["containerfile"])
	}

	pipeline, ok := doc["pipeline"].(map[string]any)
	if !ok {
		t.Fatal("pipeline missing")
	}
	rawStages, ok := pipeline["stages"].([]map[string]any)
	if !ok || len(rawStages) == 0 {
		t.Fatal("pipeline stages missing")
	}
	if rawStages[0]["id"] != "validate" {
		t.Fatalf("unexpected stage id: %#v", rawStages[0])
	}
	if _, hasScript := rawStages[0]["script"]; hasScript {
		t.Fatal("typed stages must not contain script")
	}
	if _, hasJnlp := doc["jnlp"]; hasJnlp {
		t.Fatal("manifest must not contain jnlp")
	}
	if _, hasLib := doc["lib"]; hasLib {
		t.Fatal("manifest must not contain lib section")
	}
	if _, hasExecutor := doc["executor"]; hasExecutor {
		t.Fatal("manifest must not contain executor section")
	}
	runtime, ok := doc["runtime"].(map[string]string)
	if !ok || runtime["image"] == "" {
		t.Fatalf("runtime missing: %#v", doc["runtime"])
	}
}

func TestBuilderDockerfileEngine(t *testing.T) {
	b := Builder{}
	release := sampleRelease()
	release.Content.BuildEngine = "dockerfile"
	release.Content.DockerfilePath = "Dockerfile"
	release.Content.DockerfileImageTarget = "runtime"
	release.Content.DockerfileTestTarget = "test"
	release.Content.ContainerfileKey = ""
	release.Content.ContainerfileSHA256 = ""
	release.Content.CacheRefTemplate = "{{registryHost}}/coin-cache/{{project}}:dockerfile"

	doc, _, err := b.Build(release, BuildOptions{Project: "demo-go-app", RegistryHost: "localhost:8082"})
	if err != nil {
		t.Fatal(err)
	}
	build, ok := doc["build"].(map[string]any)
	if !ok {
		t.Fatal("build missing")
	}
	if build["engine"] != "dockerfile" {
		t.Fatalf("unexpected engine: %#v", build["engine"])
	}
	df, ok := build["dockerfile"].(map[string]any)
	if !ok {
		t.Fatal("build.dockerfile missing")
	}
	if df["dockerfile"] != "Dockerfile" {
		t.Fatalf("unexpected dockerfile path: %#v", df["dockerfile"])
	}
	if df["imageTarget"] != "runtime" || df["testTarget"] != "test" {
		t.Fatalf("unexpected targets: %#v", df)
	}
	cacheRef, _ := df["cacheRef"].(string)
	if cacheRef != "localhost:8082/coin-cache/demo-go-app:dockerfile" {
		t.Fatalf("unexpected cacheRef: %q", cacheRef)
	}
	if _, hasContainerfile := df["containerfile"]; hasContainerfile {
		t.Fatal("BYO dockerfile manifest must not include containerfile ref")
	}
	if _, hasBuildkit := build["buildkit"]; hasBuildkit {
		t.Fatal("dockerfile manifest must not include buildkit section")
	}
}

func sampleRelease() GPRelease {
	return GPRelease{
		Name:    "go-app",
		Version: "1.0.0",
		Parts: Composition{
			AgentImage:       "localhost:8082/coin-docker/coin-agent:1.0.0",
			AgentDigest:      "sha256:deadbeef",
			GPContentName:    "go-app",
			GPContentVersion: "1.0.0",
		},
		Content: ContentBundle{
			SchemaArtifactKey:     "schemas/config.v2.schema.json",
			SchemaSHA256:          "sha256:schema",
			ContainerfileKey:      "dockerfiles/Containerfile",
			ContainerfileSHA256:   "sha256:containerfile",
			BuildEngine:           "buildkit",
			BuildkitTargets: map[string]string{
				"validate": "validate",
				"test":     "test",
				"image":    "runtime",
				"artifact": "artifact",
			},
			CacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:buildkit",
			Stages: []TypedStage{
				{ID: "validate", Name: "Validate"},
				{ID: "test", Name: "Test"},
				{ID: "build", Name: "Build"},
				{ID: "publish", Name: "Publish"},
			},
		},
	}
}
