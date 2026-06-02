package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

const validConfig = `
version: 1

coin:
  template: python-uv
  templateVersion: "1.0.0"

agent:
  stack: python-uv
  runtime:
    python: "3.13"
  publishRegistry: nexus-docker

project:
  name: my-service

pipeline:
  test:
    enabled: true
  build:
    target: container
  publish:
    when: tag
`

func TestLoad_Valid(t *testing.T) {
	path := writeConfig(t, validConfig)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Project.Name != "my-service" {
		t.Errorf("project.name = %q, want my-service", cfg.Project.Name)
	}
	if cfg.Agent.Stack != "python-uv" {
		t.Errorf("agent.stack = %q, want python-uv", cfg.Agent.Stack)
	}
	if cfg.Agent.Runtime["python"] != "3.13" {
		t.Errorf("agent.runtime.python = %q, want 3.13", cfg.Agent.Runtime["python"])
	}
	if cfg.Agent.PublishRegistry != "nexus-docker" {
		t.Errorf("agent.publishRegistry = %q, want nexus-docker", cfg.Agent.PublishRegistry)
	}
	if cfg.BuildTarget() != "container" {
		t.Errorf("BuildTarget = %q, want container", cfg.BuildTarget())
	}
}

func TestLoad_DefaultBuildTarget(t *testing.T) {
	path := writeConfig(t, `
version: 1
agent:
  stack: go
project:
  name: svc
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BuildTarget() != "package" {
		t.Errorf("default BuildTarget = %q, want package", cfg.BuildTarget())
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeConfig(t, "{ invalid yaml :")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoad_WrongVersion(t *testing.T) {
	path := writeConfig(t, `
version: 99
agent:
  stack: go
project:
  name: svc
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
}

func TestLoad_MissingProjectName(t *testing.T) {
	path := writeConfig(t, `
version: 1
agent:
  stack: go
project:
  name: ""
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing project.name")
	}
}

func TestLoad_MissingStack(t *testing.T) {
	path := writeConfig(t, `
version: 1
agent: {}
project:
  name: my-service
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing agent.stack")
	}
}

func TestLoad_UnknownStack(t *testing.T) {
	path := writeConfig(t, `
version: 1
agent:
  stack: rust
project:
  name: my-service
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown stack")
	}
}

func TestLoad_AllStacks(t *testing.T) {
	stacks := []string{"python-uv", "python-pip", "java-maven", "java-gradle", "go", "node"}
	for _, stack := range stacks {
		t.Run(stack, func(t *testing.T) {
			path := writeConfig(t, "version: 1\nagent:\n  stack: "+stack+"\nproject:\n  name: svc\n")
			_, err := Load(path)
			if err != nil {
				t.Errorf("stack %q: unexpected error: %v", stack, err)
			}
		})
	}
}

func TestStageIsEnabled(t *testing.T) {
	trueVal := true
	falseVal := false

	cases := []struct {
		s    Stage
		want bool
	}{
		{Stage{Enabled: nil}, true},   // nil → включена по умолчанию
		{Stage{Enabled: &trueVal}, true},
		{Stage{Enabled: &falseVal}, false},
	}
	for _, c := range cases {
		got := c.s.IsEnabled()
		if got != c.want {
			t.Errorf("IsEnabled(%v) = %v, want %v", c.s.Enabled, got, c.want)
		}
	}
}

func TestRuntimeVersion(t *testing.T) {
	cfg := &Config{
		Agent: Agent{
			Runtime: map[string]string{"python": "3.12"},
		},
	}
	if v := cfg.RuntimeVersion("python", "3.13"); v != "3.12" {
		t.Errorf("got %q, want 3.12", v)
	}
	if v := cfg.RuntimeVersion("java", "17"); v != "17" {
		t.Errorf("got %q, want 17 (default)", v)
	}
}
