package deliverables

import "testing"

func TestNormalizeDefault(t *testing.T) {
	got := Normalize(nil)
	spec, ok := got[DefaultName]
	if !ok || spec.Type != "image" || spec.Context != "." {
		t.Fatalf("unexpected default: %+v", got)
	}
}

func TestValidateDependsOn(t *testing.T) {
	items := map[string]Spec{
		"app": {Type: "image", Context: "."},
		"meta": {
			Type:      "artifact",
			Source:    ".coin/out",
			DependsOn: "missing",
		},
	}
	if err := Validate(items, P0Types); err == nil {
		t.Fatal("expected dependsOn error")
	}
}

func TestValidateArtifactGlobRejected(t *testing.T) {
	items := map[string]Spec{
		"meta": {
			Type: "artifact",
			Sources: []ArtifactSource{
				{Path: "docs/*.yaml"},
			},
		},
	}
	if err := Validate(items, P0Types); err == nil {
		t.Fatal("expected glob rejection")
	}
}
