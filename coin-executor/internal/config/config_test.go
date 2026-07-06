package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadV2(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `coin:
  goldenPath: go-app
  version: "1.0.0"
project:
  name: demo
  groupId: com.example.team
  artifactId: demo
jenkins:
  credentials:
    docker: nexus-docker
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Coin.GoldenPath != "go-app" || cfg.Coin.Version != "1.0.0" {
		t.Fatalf("unexpected coin meta: %+v", cfg.Coin)
	}
	if cfg.Project.GroupID != "com.example.team" || cfg.Project.ArtifactID != "demo" {
		t.Fatalf("unexpected project identity: %+v", cfg.Project)
	}
}

func TestLoadRejectsProductBuildPublishFields(t *testing.T) {
	tests := map[string]string{
		"project repository": `coin:
  goldenPath: go-app
  version: "1.0.0"
project:
  name: demo
  groupId: com.example.team
  artifactId: demo
  repository: maven-releases
jenkins:
  credentials:
    docker: nexus-docker
`,
		"deliverables": `coin:
  goldenPath: go-app
  version: "1.0.0"
project:
  name: demo
  groupId: com.example.team
  artifactId: demo
jenkins:
  credentials:
    docker: nexus-docker
deliverables:
  app:
    type: image
`,
		"pipeline commands": `coin:
  goldenPath: go-app
  version: "1.0.0"
project:
  name: demo
  groupId: com.example.team
  artifactId: demo
jenkins:
  credentials:
    docker: nexus-docker
pipeline:
  build:
    commands:
      - go build ./...
`,
		"build block": `coin:
  goldenPath: go-app
  version: "1.0.0"
project:
  name: demo
  groupId: com.example.team
  artifactId: demo
jenkins:
  credentials:
    docker: nexus-docker
build:
  cacheRef: localhost:8082/cache/demo
`,
	}

	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil {
				t.Fatalf("expected config to be rejected")
			}
		})
	}
}
