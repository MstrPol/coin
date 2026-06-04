package goldenpaths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"coin.local/coin-cli/internal/platform"
)

func goldenPathsRoot(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv("COIN_GOLDEN_PATHS_DIR"); dir != "" {
		if _, err := os.Stat(filepath.Join(dir, catalogFile)); err == nil {
			return dir
		}
	}
	if dir, err := platform.GoldenPathsDir(); err == nil {
		if _, err := os.Stat(filepath.Join(dir, catalogFile)); err == nil {
			return dir
		}
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repo := filepath.Join(filepath.Dir(file), "..", "..", "..")
	root := filepath.Join(repo, "coin-platform", "golden-paths")
	if _, err := os.Stat(filepath.Join(root, catalogFile)); err != nil {
		t.Fatalf("golden-paths not found at %s: %v", root, err)
	}
	return root
}

func TestResolve_pythonUvApp(t *testing.T) {
	t.Setenv("COIN_GOLDEN_PATHS_SOURCE", "local")
	t.Setenv("COIN_GOLDEN_PATHS_DIR", goldenPathsRoot(t))

	bundle, err := Resolve("python-uv-app", "v1")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if bundle.Name != "python-uv-app" {
		t.Errorf("Name = %q, want python-uv-app", bundle.Name)
	}
	if bundle.Stack() != "python-uv" {
		t.Errorf("Stack = %q, want python-uv", bundle.Stack())
	}
	if bundle.BuildType() != "container" {
		t.Errorf("BuildType = %q, want container", bundle.BuildType())
	}

	script, err := bundle.Script("test")
	if err != nil {
		t.Fatalf("Script(test): %v", err)
	}
	if script == "" {
		t.Fatal("test script is empty")
	}

	df, err := bundle.Dockerfile()
	if err != nil {
		t.Fatalf("Dockerfile: %v", err)
	}
	if df == "" {
		t.Fatal("Dockerfile is empty")
	}
	if strings.Contains(df, "AS builder") {
		t.Error("Dockerfile must be runtime-only (no AS builder); see docs/agent-build-model.md")
	}
}

func TestAllAppDockerfiles_runtimeOnly(t *testing.T) {
	t.Setenv("COIN_GOLDEN_PATHS_SOURCE", "local")
	t.Setenv("COIN_GOLDEN_PATHS_DIR", goldenPathsRoot(t))

	for _, name := range []string{
		"go-app", "python-uv-app", "python-pip-app",
		"java-gradle-app", "java-maven-app",
	} {
		t.Run(name, func(t *testing.T) {
			bundle, err := Resolve(name, "v1")
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			df, err := bundle.Dockerfile()
			if err != nil {
				t.Fatalf("Dockerfile: %v", err)
			}
			if strings.Contains(df, "AS builder") {
				t.Errorf("%s/v1 Dockerfile must be runtime-only", name)
			}
		})
	}
}

func TestResolve_unknownTemplate(t *testing.T) {
	t.Setenv("COIN_GOLDEN_PATHS_SOURCE", "local")
	t.Setenv("COIN_GOLDEN_PATHS_DIR", goldenPathsRoot(t))

	_, err := Resolve("python-uv", "v1")
	if err == nil {
		t.Fatal("expected error for unknown golden path name")
	}
}

func TestResolve_defaultVersion(t *testing.T) {
	t.Setenv("COIN_GOLDEN_PATHS_SOURCE", "local")
	t.Setenv("COIN_GOLDEN_PATHS_DIR", goldenPathsRoot(t))

	bundle, err := Resolve("go-app", "")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if bundle.Version != "v1" {
		t.Errorf("Version = %q, want v1 (latest)", bundle.Version)
	}
}

func TestLoadCatalog(t *testing.T) {
	t.Setenv("COIN_GOLDEN_PATHS_SOURCE", "local")
	t.Setenv("COIN_GOLDEN_PATHS_DIR", goldenPathsRoot(t))

	catalog, err := LoadCatalog(os.DirFS(goldenPathsRoot(t)))
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	if len(catalog.Paths) == 0 {
		t.Fatal("expected paths in catalog")
	}
}

func TestCatalog_DeprecationWarning(t *testing.T) {
	catalog := &Catalog{
		Paths: map[string]CatalogEntry{
			"python-pip-app": {
				Latest:     "v2",
				Deprecated: []string{"v0", "v1"},
			},
			"go-app": {Latest: "v1"},
		},
	}

	if got := catalog.DeprecationWarning("python-pip-app", "v1"); got != "версия v1 снята с поддержки — перейдите на v2" {
		t.Errorf("deprecated version: got %q", got)
	}
	if got := catalog.DeprecationWarning("go-app", "v1"); got != "" {
		t.Errorf("active version: got %q, want empty", got)
	}
}
