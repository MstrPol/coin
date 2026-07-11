package gpcontent

import "testing"

func TestContentBundleFromInlineDoc(t *testing.T) {
	doc, err := DefaultPipelineDoc("go-app")
	if err != nil {
		t.Fatal(err)
	}
	bundle := ContentBundleFromInlineDoc(doc, "go-app", "1.0.0")
	if len(bundle.Stages) == 0 {
		t.Fatal("expected pipeline stages")
	}
	if !bundle.IsInlinePipeline() {
		t.Fatal("expected inline pipeline bundle")
	}
}

func TestDefaultPipelineDocGoAppDocker(t *testing.T) {
	doc, err := DefaultPipelineDoc("go-app-docker")
	if err != nil {
		t.Fatal(err)
	}
	if doc.SchemaVersion != SchemaVersionInline {
		t.Fatalf("schemaVersion=%d", doc.SchemaVersion)
	}
}
