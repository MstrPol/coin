package manifest

import "testing"

func TestBuilderBranchingSection(t *testing.T) {
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
	doc, _, err := b.Build(release, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	branching, ok := doc["branching"].(map[string]any)
	if !ok {
		t.Fatalf("branching missing: %#v", doc["branching"])
	}
	if branching["name"] != "trunk-based" || branching["version"] != "1.0.0" {
		t.Fatalf("branching=%#v", branching)
	}
}
