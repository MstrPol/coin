package scanner

import (
	"errors"
	"testing"
)

func TestParseConfigV2(t *testing.T) {
	raw := []byte(`
coin:
  goldenPath: go-app
  version: "1.0.0"
project:
  name: demo-go-app
`)
	p, err := ParseConfig(raw, "fallback")
	if err != nil {
		t.Fatal(err)
	}
	if p.Project != "demo-go-app" || p.GoldenPath != "go-app" || p.Version != "1.0.0" {
		t.Fatalf("%+v", p)
	}
}

func TestParseConfigV1Rejected(t *testing.T) {
	raw := []byte(`
coin:
  template: python-pip-app
  templateVersion: v1
`)
	_, err := ParseConfig(raw, "demo")
	if !errors.Is(err, ErrConfigV1) {
		t.Fatalf("expected ErrConfigV1, got %v", err)
	}
}

func TestParseConfigFallbackName(t *testing.T) {
	raw := []byte(`coin: { goldenPath: go-app, version: "1.0.0" }`)
	p, err := ParseConfig(raw, "my-repo")
	if err != nil {
		t.Fatal(err)
	}
	if p.Project != "my-repo" {
		t.Fatalf("project %q", p.Project)
	}
}

func TestParseConfigInvalidYAML(t *testing.T) {
	_, err := ParseConfig([]byte(":\n  bad"), "x")
	if !errors.Is(err, ErrBadConfig) {
		t.Fatalf("%v", err)
	}
}

func TestParseConfigEmptyGP(t *testing.T) {
	_, err := ParseConfig([]byte("coin: {}\n"), "x")
	if !errors.Is(err, ErrBadConfig) {
		t.Fatalf("%v", err)
	}
}

func TestSplitOwnerRepo(t *testing.T) {
	o, n, err := SplitOwnerRepo("coin/demo-go-app")
	if err != nil || o != "coin" || n != "demo-go-app" {
		t.Fatalf("%s %s %v", o, n, err)
	}
	_, _, err = SplitOwnerRepo("bad")
	if err == nil {
		t.Fatal("expected error")
	}
}
