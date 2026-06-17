package executor

import (
	"net/http"
	"net/http/httptest"
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

func TestRunValidateStage(t *testing.T) {
	policySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"warning":""}`))
	}))
	t.Cleanup(policySrv.Close)

	root := t.TempDir()
	t.Setenv("COIN_API_URL", policySrv.URL)
	t.Setenv("COIN_API_TOKEN", "test-token")
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
  repository: maven-releases
deliverables:
  app:
    type: image
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	containerfile := filepath.Join(root, ".coin", "Containerfile")
	if err := os.WriteFile(containerfile, []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(root, ".coin", "manifest.json")
	m := `{
  "manifestVersion": 1,
  "goldenPath": {"name": "go-app", "version": "1.0.0"},
  "executor": {"version": "0.1.0", "url": "http://example/executor"},
  "runtime": {"image": "coin-agent:1.0.0"},
  "build": {
    "engine": "buildkit",
    "buildkit": {
      "dockerfile": ".coin/Containerfile",
      "targets": {"validate": "validate", "test": "test", "image": "runtime"},
      "containerfile": {"url": "http://example/containerfile", "sha256": ""}
    }
  },
  "pipeline": {
    "stages": [
      {"id": "validate", "name": "Validate"},
      {"id": "test", "name": "Test"}
    ]
  },
  "validateSchema": {"url": "http://example/schema", "sha256": ""},
  "capabilities": {"deliverables": ["image"]}
}`
	if err := os.WriteFile(manifestPath, []byte(m), 0o644); err != nil {
		t.Fatal(err)
	}

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
