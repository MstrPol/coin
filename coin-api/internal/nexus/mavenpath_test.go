package nexus

import "testing"

func TestClassifierFromArtifactKey(t *testing.T) {
	if got := ClassifierFromArtifactKey("scripts/test.sh"); got != "scripts.test" {
		t.Fatalf("got %q", got)
	}
	if got := ClassifierFromArtifactKey("schemas/config.v2.schema.json"); got != "schemas.config.v2.schema" {
		t.Fatalf("got %q", got)
	}
	if got := ClassifierFromArtifactKey("dockerfiles/Containerfile"); got != "dockerfiles.Containerfile" {
		t.Fatalf("got %q", got)
	}
}

func TestArtifactMavenCoords(t *testing.T) {
	c, e := ArtifactMavenCoords("dockerfiles/Containerfile")
	if c != "dockerfiles.Containerfile" || e != "containerfile" {
		t.Fatalf("Containerfile coords: %q %q", c, e)
	}
	c, e = ArtifactMavenCoords("schemas/config.v2.schema.json")
	if c != "schemas.config.v2.schema" || e != "json" {
		t.Fatalf("schema coords: %q %q", c, e)
	}
}

func TestImmutableConflict(t *testing.T) {
	body := "cannot be updated as asset already exists and redeploy is not allowed"
	if !ImmutableConflict(400, body) {
		t.Fatal("expected immutable conflict")
	}
	status := "400 maven-releases/coin/manifest/go-app/1.0.0/go-app-1.0.0.json cannot be updated as asset already exists and redeploy is not allowed"
	if !ImmutableConflict(400, status) {
		t.Fatal("expected conflict from status line text")
	}
	if ImmutableConflict(404, body) {
		t.Fatal("expected no conflict for 404")
	}
	if ImmutableConflict(400, "not found") {
		t.Fatal("unexpected conflict")
	}
}
