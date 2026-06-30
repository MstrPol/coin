package gpcontent

import (
	"testing"
)

func TestValidateDoc_buildkit(t *testing.T) {
	doc := Doc{
		SchemaVersion: 2,
		Name:          "go-app",
		Kind:          "gp-content",
		Capabilities:  Capabilities{Deliverables: []string{"image", "artifact"}},
		Build: Build{
			Engine: "buildkit",
			Buildkit: &BuildkitBlock{
				Targets:          map[string]string{"test": "test"},
				CacheRefTemplate: "{{registryHost}}/cache/{{project}}",
			},
		},
		Pipeline: Pipeline{Stages: []Stage{{ID: "test", Name: "Test"}}},
		Artifacts: Artifacts{
			ValidateSchema: "schemas/config.v2.schema.json",
			Containerfile:  "dockerfiles/Containerfile",
		},
	}
	issues, _ := ValidateDoc(doc, PreviewOptions{ComponentName: "go-app", HasContainerfileArtifact: true})
	if len(issues) != 0 {
		t.Fatalf("expected valid, got %v", issues)
	}
}

func TestValidateDoc_rejectBuildpack(t *testing.T) {
	doc := Doc{SchemaVersion: 2, Name: "x", Kind: "gp-content", Build: Build{Engine: "buildpack"}}
	issues, _ := ValidateDoc(doc, PreviewOptions{ComponentName: "x"})
	found := false
	for _, iss := range issues {
		if iss.Field == "content.yaml.build.engine" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected buildpack rejection")
	}
}

func TestValidateDoc_byoRejectsArtifact(t *testing.T) {
	doc := Doc{
		SchemaVersion: 2,
		Name:          "go-app-docker",
		Kind:          "gp-content",
		Capabilities:  Capabilities{Deliverables: []string{"image", "artifact"}},
		Build: Build{
			Engine: "dockerfile",
			Dockerfile: &DockerfileBlock{
				Path:             "Dockerfile",
				CacheRefTemplate: "{{registryHost}}/cache",
			},
		},
		Pipeline:  Pipeline{Stages: []Stage{{ID: "build", Name: "Build"}}},
		Artifacts: Artifacts{ValidateSchema: "schemas/config.v2.schema.json"},
	}
	issues, _ := ValidateDoc(doc, PreviewOptions{ComponentName: "go-app-docker"})
	if len(issues) == 0 {
		t.Fatal("expected artifact rejection for BYO engine")
	}
}

func TestPreview_byoNoContainerfileRef(t *testing.T) {
	doc := Doc{
		SchemaVersion: 2,
		Name:          "go-app-docker",
		Kind:          "gp-content",
		Capabilities:  Capabilities{Deliverables: []string{"image"}},
		Build: Build{
			Engine: "dockerfile",
			Dockerfile: &DockerfileBlock{
				Path:             "Dockerfile",
				ImageTarget:      "runtime",
				CacheRefTemplate: "{{registryHost}}/coin-cache/{{project}}:dockerfile",
			},
		},
		Pipeline:  Pipeline{Stages: []Stage{{ID: "build", Name: "Build"}}},
		Artifacts: Artifacts{ValidateSchema: "schemas/config.v2.schema.json"},
	}
	res := Preview(doc, PreviewOptions{ComponentName: "go-app-docker", Project: "demo"})
	if !res.Valid {
		t.Fatalf("preview invalid: %v", res.Issues)
	}
	df, ok := res.Build["dockerfile"].(map[string]any)
	if !ok {
		t.Fatal("expected dockerfile block")
	}
	if df["dockerfile"] != "Dockerfile" {
		t.Fatalf("unexpected dockerfile path %v", df["dockerfile"])
	}
	if _, has := df["containerfile"]; has {
		t.Fatal("BYO preview must not include containerfile ref")
	}
}
