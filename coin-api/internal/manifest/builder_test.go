package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuilderStableHash(t *testing.T) {
	b := Builder{}
	release := sampleRelease()
	opts := BuildOptions{}
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
	doc, _, err := b.Build(sampleRelease(), BuildOptions{})
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
	if _, hasCacheRef := buildkit["cacheRef"]; hasCacheRef {
		t.Fatal("build.buildkit must not contain cacheRef")
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
	if _, hasCredentials := doc["credentials"]; hasCredentials {
		t.Fatal("manifest must not contain Jenkins credentials")
	}
	runtime, ok := doc["runtime"].(map[string]string)
	if !ok || runtime["image"] == "" {
		t.Fatalf("runtime missing: %#v", doc["runtime"])
	}
	dest, ok := doc["destinations"].(map[string]any)
	if !ok || dest["imageRegistryPrefix"] == "" || dest["artifactRepositoryBase"] == "" {
		t.Fatalf("destinations missing: %#v", doc["destinations"])
	}
}

func TestBuilderManifestTopLevelShape(t *testing.T) {
	b := Builder{}
	release := sampleRelease()
	release.Branching = BranchingBundle{
		Name:    "trunk-based",
		Version: "1.0.0",
		Rules: map[string]any{
			"branches": []any{
				map[string]any{
					"name":       "main",
					"pattern":    `^main$`,
					"versioning": map[string]any{"template": "v{base}-main-snapshot-{n}"},
					"publish":    false,
				},
			},
		},
	}
	release.Content.Capabilities = map[string]any{
		"deliverables": []any{"image"},
	}

	doc, _, err := b.Build(release, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}

	allowed := map[string]bool{
		"manifestVersion": true,
		"manifestHash":    true,
		"goldenPath":      true,
		"runtime":         true,
		"destinations":    true,
		"build":           true,
		"pipeline":        true,
		"validateSchema":  true,
		"capabilities":    true,
		"branching":       true,
	}
	for key := range doc {
		if !allowed[key] {
			t.Fatalf("manifest contains non-composition top-level key %q", key)
		}
	}
	for _, key := range []string{"goldenPath", "runtime", "destinations", "build", "pipeline", "validateSchema", "capabilities", "branching", "manifestVersion", "manifestHash"} {
		if _, ok := doc[key]; !ok {
			t.Fatalf("manifest missing required materialized key %q", key)
		}
	}
}

func TestManifestSchemaRejectsCredentialsField(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "manifest.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	if _, ok := schema.Properties["credentials"]; ok {
		t.Fatal("manifest schema must not allow top-level credentials")
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

	doc, _, err := b.Build(release, BuildOptions{})
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
	if _, hasCacheRef := df["cacheRef"]; hasCacheRef {
		t.Fatal("build.dockerfile must not contain cacheRef")
	}
	if _, hasContainerfile := df["containerfile"]; hasContainerfile {
		t.Fatal("BYO dockerfile manifest must not include containerfile ref")
	}
	if _, hasBuildkit := build["buildkit"]; hasBuildkit {
		t.Fatal("dockerfile manifest must not include buildkit section")
	}
}

func TestBuilderVNextBuildContract(t *testing.T) {
	b := Builder{}
	doc, _, err := b.Build(sampleVNextRelease(), BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if _, hasCapabilities := doc["capabilities"]; hasCapabilities {
		t.Fatal("vNext manifest must not use flat capabilities deliverables")
	}
	parameters, ok := doc["parameters"].([]Parameter)
	if !ok || len(parameters) != 1 || parameters[0].Name != "GO_VERSION" {
		t.Fatalf("parameters missing: %#v", doc["parameters"])
	}
	deliverables, ok := doc["deliverables"].([]Deliverable)
	if !ok || len(deliverables) != 2 || deliverables[0].ID != "app" {
		t.Fatalf("deliverables missing: %#v", doc["deliverables"])
	}
	artifacts, ok := doc["artifacts"].(map[string]any)
	if !ok {
		t.Fatalf("artifacts missing: %#v", doc["artifacts"])
	}
	containerfiles, ok := artifacts["containerfiles"].([]map[string]string)
	if !ok || len(containerfiles) != 1 || containerfiles[0]["id"] != "app" {
		t.Fatalf("containerfile refs missing: %#v", artifacts["containerfiles"])
	}
	build, ok := doc["build"].(map[string]any)
	if !ok {
		t.Fatal("build missing")
	}
	if _, hasEngine := build["engine"]; hasEngine {
		t.Fatal("vNext build must not contain top-level engine")
	}
	targets, ok := build["targets"].([]BuildTarget)
	if !ok || len(targets) != 2 || targets[0].ID != "app-image" {
		t.Fatalf("targets missing: %#v", build["targets"])
	}
	pipeline, ok := doc["pipeline"].(map[string]any)
	if !ok {
		t.Fatal("pipeline missing")
	}
	stages, ok := pipeline["stages"].([]map[string]any)
	if !ok || len(stages) != 1 {
		t.Fatalf("stages missing: %#v", pipeline["stages"])
	}
	steps, ok := stages[0]["steps"].([]StageStep)
	if !ok || len(steps) != 2 || steps[0].DeliverableID != "app" {
		t.Fatalf("stage steps missing: %#v", stages[0]["steps"])
	}
}

func TestBuilderVNextHashIncludesTargets(t *testing.T) {
	b := Builder{}
	release := sampleVNextRelease()
	_, hash1, err := b.Build(release, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	release.Content.Targets[0].Target = "release"
	_, hash2, err := b.Build(release, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if hash1 == hash2 {
		t.Fatal("manifest hash must change when vNext target changes")
	}
}

func sampleRelease() GPRelease {
	return GPRelease{
		Name:    "go-app",
		Version: "1.0.0",
		Destinations: Destinations{
			ImageRegistryPrefix:    "localhost:8082/coin-docker",
			BuildCacheEnabled:      true,
			ArtifactRepositoryBase: "http://nexus:8081/repository/maven-releases",
		},
		Parts: Composition{
			AgentImage:       "localhost:8082/coin-docker/coin-agent:1.0.0",
			AgentDigest:      "sha256:deadbeef",
			GPContentName:    "go-app",
			GPContentVersion: "1.0.0",
		},
		Content: ContentBundle{
			SchemaArtifactKey:   "schemas/config.v2.schema.json",
			SchemaSHA256:        "sha256:schema",
			ContainerfileKey:    "dockerfiles/Containerfile",
			ContainerfileSHA256: "sha256:containerfile",
			BuildEngine:         "buildkit",
			BuildkitTargets: map[string]string{
				"validate": "validate",
				"test":     "test",
				"image":    "runtime",
				"artifact": "artifact",
			},
			Stages: []TypedStage{
				{ID: "validate", Name: "Validate"},
				{ID: "test", Name: "Test"},
				{ID: "build", Name: "Build"},
				{ID: "publish", Name: "Publish"},
			},
		},
	}
}

func sampleVNextRelease() GPRelease {
	release := sampleRelease()
	release.Content.Capabilities = nil
	release.Content.BuildEngine = ""
	release.Content.BuildkitTargets = nil
	release.Content.ContainerfileKey = ""
	release.Content.ContainerfileSHA256 = ""
	release.Content.Parameters = []Parameter{
		{Name: "GO_VERSION", Type: "string", Default: "1.22", Required: true},
	}
	release.Content.Targets = []BuildTarget{
		{ID: "app-image", Engine: "buildkit", Containerfile: "app", Target: "runtime"},
		{ID: "app-artifact", Engine: "buildkit", Containerfile: "app", Target: "artifact"},
	}
	release.Content.Deliverables = []Deliverable{
		{ID: "app", Type: "image", TargetID: "app-image"},
		{ID: "app-zip", Type: "artifact", TargetID: "app-artifact", Artifact: &ArtifactDeliverable{Format: "zip", Paths: []string{"/out/app"}}},
	}
	release.Content.Containerfiles = []NamedContentRef{
		{ID: "app", ArtifactKey: "dockerfiles/app.Containerfile", SHA256: "sha256:containerfile"},
	}
	release.Content.Stages = []TypedStage{
		{
			ID:   "build",
			Name: "Build",
			Steps: []StageStep{
				{Action: "build-deliverable", DeliverableID: "app"},
				{Action: "build-deliverable", DeliverableID: "app-zip"},
			},
		},
	}
	return release
}
