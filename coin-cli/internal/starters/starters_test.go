package starters

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func startersRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repo := filepath.Join(filepath.Dir(file), "..", "..", "..")
	root := filepath.Join(repo, "coin-platform", "starters")
	if !isStarterRoot(root) {
		t.Fatalf("starters not found at %s", root)
	}
	return root
}

func TestList(t *testing.T) {
	root := startersRoot(t)
	names, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 5 {
		t.Fatalf("expected at least 5 starters, got %v", names)
	}
	found := false
	for _, n := range names {
		if n == "python-uv-app" {
			found = true
		}
	}
	if !found {
		t.Fatalf("python-uv-app not in %v", names)
	}
}

func TestMaterialize(t *testing.T) {
	root := startersRoot(t)
	dest := t.TempDir()

	p := Params{
		Starter:     "go-app",
		ProjectName: "demo-svc",
		GroupID:     "com.test",
		Repository:  "Nexus_PROD",
		DockerCred:  "nexus-docker",
		DestDir:     dest,
	}
	if err := Materialize(root, p); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(dest, ".coin", "config.yaml")
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "demo-svc") {
		t.Fatalf("config missing project name: %s", raw)
	}
	if _, err := os.Stat(filepath.Join(dest, "main.go")); err != nil {
		t.Fatal("main.go not copied")
	}
}
