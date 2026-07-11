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
				Targets: map[string]string{"test": "test"},
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
				Path: "Dockerfile",
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

func TestValidateDoc_rejectsInvalidDeliverables(t *testing.T) {
	doc := Doc{
		SchemaVersion: 2,
		Name:          "go-app",
		Kind:          "gp-content",
		Capabilities:  Capabilities{Deliverables: []string{"image", "image", "wheel"}},
		Build: Build{
			Engine:   "buildkit",
			Buildkit: &BuildkitBlock{Targets: map[string]string{"image": "runtime"}},
		},
		Pipeline: Pipeline{Stages: []Stage{{ID: "build", Name: "Build"}}},
		Artifacts: Artifacts{
			ValidateSchema: "schemas/config.v2.schema.json",
			Containerfile:  "dockerfiles/Containerfile",
		},
	}
	issues, _ := ValidateDoc(doc, PreviewOptions{ComponentName: "go-app", HasContainerfileArtifact: true})
	if len(issues) == 0 {
		t.Fatal("expected invalid deliverables to be rejected")
	}
}

func TestValidateDoc_vNextCatalogSupersededFromLegacyHelper(t *testing.T) {
	doc := validVNextDoc()
	issues, _ := ValidateDoc(doc, PreviewOptions{ComponentName: "go-app"})
	if len(issues) == 0 {
		t.Fatal("expected v2 catalog superseded rejection")
	}
	assertIssueField(t, issues, "content.yaml.schemaVersion")
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
				Path:        "Dockerfile",
				ImageTarget: "runtime",
			},
		},
		Pipeline:  Pipeline{Stages: []Stage{{ID: "build", Name: "Build"}}},
		Artifacts: Artifacts{ValidateSchema: "schemas/config.v2.schema.json"},
	}
	res := Preview(doc, PreviewOptions{ComponentName: "go-app-docker"})
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
	if _, has := df["cacheRef"]; has {
		t.Fatal("BYO preview must not include cacheRef")
	}
}

func validVNextDoc() Doc {
	return Doc{
		SchemaVersion: 2,
		Name:          "go-app",
		Kind:          "gp-content",
		Parameters: []Parameter{
			{Name: "GO_VERSION", Type: "string", Default: "1.22", Required: true},
		},
		Build: Build{
			Targets: []BuildTarget{
				{ID: "app-image", Engine: "buildkit", Containerfile: "app", Target: "runtime"},
				{ID: "app-artifact", Engine: "buildkit", Containerfile: "app", Target: "artifact"},
			},
		},
		Deliverables: []Deliverable{
			{ID: "app", Type: "image", TargetID: "app-image"},
			{ID: "app-zip", Type: "artifact", TargetID: "app-artifact"},
		},
		Pipeline: Pipeline{Stages: []Stage{
			{
				ID:   "build",
				Name: "Build",
				Steps: []StageStep{
					{Action: "build-deliverable", DeliverableID: "app"},
					{Action: "build-deliverable", DeliverableID: "app-zip"},
				},
			},
		}},
		Artifacts: Artifacts{
			ValidateSchema: "schemas/config.v2.schema.json",
			Containerfiles: []ContainerfileSpec{
				{ID: "app", Path: "dockerfiles/app.Containerfile"},
			},
		},
	}
}

func assertIssueField(t *testing.T, issues []Issue, field string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Field == field {
			return
		}
	}
	t.Fatalf("expected issue field %q in %v", field, issues)
}
