package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func platformRoot(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv("COIN_PLATFORM_DIR"); dir != "" {
		if got, err := Root(); err == nil {
			return got
		}
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repo := filepath.Join(filepath.Dir(file), "..", "..", "..")
	root := filepath.Join(repo, "coin-platform")
	t.Setenv("COIN_PLATFORM_DIR", root)
	return root
}

func TestValidate_ok(t *testing.T) {
	platformRoot(t)
	if err := Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRoot_fromEnv(t *testing.T) {
	root := platformRoot(t)
	got, err := Root()
	if err != nil {
		t.Fatal(err)
	}
	if got != root {
		t.Fatalf("Root()=%q want %q", got, root)
	}
}

func TestGoldenPathsDir(t *testing.T) {
	root := platformRoot(t)
	dir, err := GoldenPathsDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "golden-paths")
	if dir != want {
		t.Fatalf("GoldenPathsDir()=%q want %q", dir, want)
	}
}

func TestRoot_missingEnv(t *testing.T) {
	t.Setenv("COIN_PLATFORM_DIR", "")
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if _, err := Root(); err == nil {
		t.Fatal("expected error without platform dir")
	}
}
