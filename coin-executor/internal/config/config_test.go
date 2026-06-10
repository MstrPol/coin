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
}
