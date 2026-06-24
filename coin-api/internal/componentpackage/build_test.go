package componentpackage

import "testing"

func TestBuildPackageManifestJSON(t *testing.T) {
	t.Parallel()
	raw, err := BuildPackageManifestJSON("branching-model", "trunk-based", "1.0.0", []ArtifactInput{
		{Path: "model.yaml", SHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{Path: "schema/rules.json", SHA256: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
	})
	if err != nil {
		t.Fatal(err)
	}
	m, err := ValidatePackageManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Files) != 2 || m.Files[0].Role != "primary" || m.Files[1].Role != "schema" {
		t.Fatalf("unexpected manifest: %+v", m)
	}
}

func TestInferFileRole(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"model.yaml":              "primary",
		"dockerfiles/Containerfile": "containerfile",
		"schema/config.v2.schema.json": "schema",
		"README.md":               "docs",
		"package.zip":             "archive",
		"scripts/validate.sh":     "other",
	}
	for path, want := range cases {
		if got := InferFileRole(path); got != want {
			t.Errorf("%s: got %q want %q", path, got, want)
		}
	}
}
