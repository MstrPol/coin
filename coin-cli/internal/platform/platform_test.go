package platform

import (
	"path/filepath"
	"runtime"
	"testing"
)

func platformRoot(t *testing.T) string {
	t.Helper()
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
	t.Chdir(t.TempDir())
	if _, err := Root(); err == nil {
		t.Fatal("expected error without platform dir")
	}
}
