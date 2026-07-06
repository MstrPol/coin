package executor

import (
	"testing"

	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/manifest"
)

func TestEnforcePublishPolicy_denied(t *testing.T) {
	m := &manifest.Manifest{
		Branching: &manifest.Branching{
			Name:    "trunk-based",
			Version: "1.0.0",
			Branches: []manifest.BranchRule{
				{Name: "feature", Pattern: `^feature/(?P<jira>[A-Z][A-Z0-9]*-\d+)(?:-.+)?$`, Versioning: manifest.BranchVersioning{Template: "v{base}-{jira}-snapshot-{n}"}, Publish: false},
			},
		},
	}
	t.Setenv("COIN_PUBLISH_REQUEST", "true")
	t.Setenv("GIT_BRANCH", "feature/PROJ-101")

	if err := enforcePublishPolicy(t.TempDir(), m); err == nil {
		t.Fatal("expected publish policy error")
	}
}

func TestEnforcePublishPolicy_allowed(t *testing.T) {
	m := &manifest.Manifest{
		Branching: &manifest.Branching{
			Name:    "trunk-based",
			Version: "1.0.0",
			Branches: []manifest.BranchRule{
				{Name: "release", Pattern: `^release/(?P<jira>[A-Z][A-Z0-9]*-\d+)(?:-.+)?$`, Versioning: manifest.BranchVersioning{Template: "v{base}-{jira}-rc-{n}"}, Publish: true},
			},
		},
	}
	t.Setenv("COIN_PUBLISH_REQUEST", "true")
	t.Setenv("GIT_BRANCH", "release/PROJ-404")

	if err := enforcePublishPolicy(t.TempDir(), m); err != nil {
		t.Fatal(err)
	}
}

func TestShouldSkipPublish_legacyTag(t *testing.T) {
	t.Setenv("TAG_NAME", "")
	m := &manifest.Manifest{
		Pipeline: manifest.Pipeline{
			Stages: []manifest.Stage{{ID: "publish", Name: "Publish", When: "tag"}},
		},
	}
	skip, _ := shouldSkipPublish(t.TempDir(), m)
	if !skip {
		t.Fatal("expected legacy skip")
	}
}

func TestImageAndCacheRefsFromDestinations(t *testing.T) {
	cfg := &config.Config{
		Project: config.Project{
			Name:       "demo-go-app",
			GroupID:    "com.example.team",
			ArtifactID: "demo-go-app",
		},
	}
	m := &manifest.Manifest{
		GoldenPath: manifest.GoldenPath{Version: "1.0.0"},
		Destinations: manifest.Destinations{
			ImageRegistryPrefix:    "docker-dev.registry.domain.ru",
			BuildCacheEnabled:      true,
			ArtifactRepositoryBase: "http://nexus:8081/repository/maven-releases",
		},
	}
	t.Setenv("COIN_IMAGE_TAG", "v1.2.3")

	wantImage := "docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app:v1.2.3"
	if got := imageRefForProject(cfg, m, t.TempDir()); got != wantImage {
		t.Fatalf("image ref = %q, want %q", got, wantImage)
	}

	wantCache := "docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app-cache"
	if got := cacheRefForProject(cfg, m); got != wantCache {
		t.Fatalf("cache ref = %q, want %q", got, wantCache)
	}

	wantLiquibase := "docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app-liquibase"
	if got := imageRepositoryForProject(cfg, m, "-liquibase"); got != wantLiquibase {
		t.Fatalf("liquibase image repository = %q, want %q", got, wantLiquibase)
	}
}

func TestCacheRefDisabled(t *testing.T) {
	cfg := &config.Config{Project: config.Project{Name: "app", GroupID: "com.example", ArtifactID: "app"}}
	m := &manifest.Manifest{
		Destinations: manifest.Destinations{
			ImageRegistryPrefix:    "docker-dev.registry.domain.ru",
			BuildCacheEnabled:      false,
			ArtifactRepositoryBase: "http://nexus:8081/repository/maven-releases",
		},
	}
	if got := cacheRefForProject(cfg, m); got != "" {
		t.Fatalf("cache ref = %q, want empty", got)
	}
}

func TestImageRepositoryUsesRuntimeRegistryOverride(t *testing.T) {
	t.Setenv("COIN_REGISTRY_PREFIX", "nexus:8082/coin-docker")
	cfg := &config.Config{Project: config.Project{Name: "app", GroupID: "com.example", ArtifactID: "app"}}
	m := &manifest.Manifest{
		Destinations: manifest.Destinations{
			ImageRegistryPrefix:    "localhost:8082/coin-docker",
			BuildCacheEnabled:      true,
			ArtifactRepositoryBase: "http://nexus:8081/repository/maven-releases",
		},
	}
	want := "nexus:8082/coin-docker/com.example/app/app-cache"
	if got := cacheRefForProject(cfg, m); got != want {
		t.Fatalf("cache ref = %q, want %q", got, want)
	}
}

func TestArtifactRepositoryURLFromDestinations(t *testing.T) {
	m := &manifest.Manifest{
		Destinations: manifest.Destinations{
			ArtifactRepositoryBase: "http://nexus:8081/repository/maven-releases/",
		},
	}
	if got, want := artifactRepositoryURL(m), "http://nexus:8081/repository/maven-releases"; got != want {
		t.Fatalf("artifact repository url = %q, want %q", got, want)
	}
}

func TestImageRefForDeliverableUsesMetadataSuffix(t *testing.T) {
	cfg := &config.Config{
		Project: config.Project{Name: "demo-go-app", GroupID: "com.example.team", ArtifactID: "demo-go-app"},
	}
	m := &manifest.Manifest{
		GoldenPath: manifest.GoldenPath{Version: "1.0.0"},
		Destinations: manifest.Destinations{
			ImageRegistryPrefix:    "docker-dev.registry.domain.ru",
			ArtifactRepositoryBase: "http://nexus:8081/repository/maven-releases",
		},
	}
	t.Setenv("COIN_IMAGE_TAG", "v2")
	got := imageRefForDeliverable(cfg, m, manifest.Deliverable{
		ID:       "liquibase",
		Type:     "image",
		TargetID: "liquibase-image",
		Image:    manifest.ImageDeliverable{RepositorySuffix: "-liquibase"},
	}, t.TempDir())
	want := "docker-dev.registry.domain.ru/com.example.team/demo-go-app/demo-go-app-liquibase:v2"
	if got != want {
		t.Fatalf("deliverable image ref = %q, want %q", got, want)
	}
}
