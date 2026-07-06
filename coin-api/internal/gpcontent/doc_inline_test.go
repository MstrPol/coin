package gpcontent

import "testing"

func validInlineDoc() Doc {
	body := "FROM golang:1.22 AS base\nWORKDIR /src\n"
	return Doc{
		SchemaVersion:  SchemaVersionInline,
		Name:           "go-app",
		Kind:           "gp-content",
		ValidateSchema: "schemas/config.v2.schema.json",
		Parameters: []Parameter{
			{Name: "GO_VERSION", Type: "string", Default: "1.22", Required: true},
		},
		Pipeline: Pipeline{Stages: []Stage{
			{
				ID:   "a3f8b2",
				Name: "Validate",
				Steps: []StageStep{{
					Action: "run",
					Run: &InlineRun{
						Engine: "buildkit",
						Output: "validate",
						Containerfile: &InlineContainerfile{Body: body},
					},
				}},
			},
			{
				ID:   "c91d4e",
				Name: "Build",
				Steps: []StageStep{
					{
						Action: "build",
						Build: &InlineBuild{
							ID:     "4e8f2a",
							Type:   "image",
							Engine: "buildkit",
							Containerfile: &InlineContainerfile{Body: body},
						},
					},
					{
						Action:  "publish",
						Publish: &InlinePublish{BuildStepID: "4e8f2a"},
					},
				},
			},
		}},
	}
}

func TestValidateDoc_inlineRejectsSemanticStageId(t *testing.T) {
	doc := validInlineDoc()
	doc.Pipeline.Stages[0].ID = "validate"
	issues, _ := ValidateDoc(doc, PreviewOptions{ComponentName: "go-app"})
	if len(issues) == 0 {
		t.Fatal("expected semantic stage id rejection")
	}
}

func TestValidateDoc_inline(t *testing.T) {
	issues, _ := ValidateDoc(validInlineDoc(), PreviewOptions{ComponentName: "go-app"})
	if len(issues) != 0 {
		t.Fatalf("expected valid inline doc, got %v", issues)
	}
}

func TestValidateDoc_inlineRejectsCatalogSections(t *testing.T) {
	doc := validInlineDoc()
	doc.Build.Targets = []BuildTarget{{ID: "x", Engine: "buildkit"}}
	issues, _ := ValidateDoc(doc, PreviewOptions{ComponentName: "go-app"})
	if len(issues) == 0 {
		t.Fatal("expected catalog sections rejection")
	}
}

func TestValidateDoc_inlineRejectsMissingContainerfileBody(t *testing.T) {
	doc := validInlineDoc()
	doc.Pipeline.Stages[0].Steps[0].Run.Containerfile.Body = ""
	issues, _ := ValidateDoc(doc, PreviewOptions{ComponentName: "go-app"})
	if len(issues) == 0 {
		t.Fatal("expected missing containerfile body rejection")
	}
}

func TestValidateDoc_vNextCatalogSuperseded(t *testing.T) {
	doc := validVNextDoc()
	issues, _ := ValidateDoc(doc, PreviewOptions{ComponentName: "go-app"})
	if len(issues) == 0 {
		t.Fatal("expected v2 catalog superseded rejection")
	}
}

func TestPreview_inline(t *testing.T) {
	doc := validInlineDoc()
	doc.Version = "1.0.0"
	res := Preview(doc, PreviewOptions{ComponentName: "go-app"})
	if !res.Valid {
		t.Fatalf("preview invalid: %v", res.Issues)
	}
	stages, ok := res.Pipeline["stages"].([]Stage)
	if !ok || len(stages) == 0 {
		t.Fatalf("stages missing: %#v", res.Pipeline)
	}
	cf := stages[0].Steps[0].Run.Containerfile
	if cf == nil || cf.Digest == "" || cf.ContentRef == nil {
		t.Fatalf("expected materialized containerfile on step: %#v", cf)
	}
	if cf.Body != "" {
		t.Fatal("preview must not include raw body after materialization")
	}
}
