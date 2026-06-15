package executor

import (
	"os"
	"path/filepath"
	"testing"

	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/manifest"
)

func TestShouldRunStage(t *testing.T) {
	t.Setenv("TAG_NAME", "")
	t.Setenv("GIT_TAG_NAME", "")

	if !shouldRunStage(manifest.Stage{Name: "test"}) {
		t.Fatal("expected always stage to run")
	}
	if shouldRunStage(manifest.Stage{Name: "publish", When: "tag"}) {
		t.Fatal("expected tag stage to skip without TAG_NAME")
	}

	t.Setenv("TAG_NAME", "v1.0.0")
	if !shouldRunStage(manifest.Stage{Name: "publish", When: "tag"}) {
		t.Fatal("expected tag stage to run with TAG_NAME")
	}
}

func TestRunSingleStage(t *testing.T) {
	root := t.TempDir()
	contentRoot := filepath.Join(root, "platform")
	gpDir := filepath.Join(contentRoot, "content", "golden-paths", "go-app", "1.0.0", "scripts")
	if err := os.MkdirAll(gpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(gpDir, "validate.sh")
	if err := os.WriteFile(script, []byte("#!/bin/bash\necho ok-validate\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	dockerfile := filepath.Join(contentRoot, "content", "golden-paths", "go-app", "1.0.0", "Dockerfile")
	if err := os.MkdirAll(filepath.Dir(dockerfile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dockerfile, []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(root, ".coin", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	cfgYAML := `coin:
  goldenPath: go-app
  version: "1.0.0"
jenkins:
  credentials:
    docker: nexus-docker
project:
  name: demo
  artifactId: demo
  groupId: com.example
  repository: Nexus_PROD
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(root, ".coin", "manifest.json")
	m := `{
  "manifestVersion": 1,
  "goldenPath": {"name": "go-app", "version": "1.0.0"},
  "executor": {"version": "0.1.0", "url": "http://example/executor"},
  "runtime": {"image": "ci-go:1.22"},
  "pipeline": {
    "stages": [
      {"name": "validate", "script": {"gitRef": "seed-local", "path": "content/golden-paths/go-app/1.0.0/scripts/validate.sh"}},
      {"name": "test", "script": {"gitRef": "seed-local", "path": "content/golden-paths/go-app/1.0.0/scripts/test.sh"}}
    ]
  },
  "validateSchema": {"gitRef": "seed-local", "path": "content/schema/config.v2.schema.json"},
  "dockerfileTemplate": {"gitRef": "seed-local", "path": "content/golden-paths/go-app/1.0.0/Dockerfile"}
}`
	if err := os.WriteFile(manifestPath, []byte(m), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COIN_CONTENT_DIR", contentRoot)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := manifest.Load(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	r := Runner{Workspace: root}
	if err := r.Run(cfg, loaded, RunOptions{Stage: "validate"}); err != nil {
		t.Fatalf("run validate: %v", err)
	}
}
