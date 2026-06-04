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
coin:
  template: python-uv-app
  templateVersion: v1

jenkins:
  runtime:
    python: "3.13"
  credentials:
    docker: nexus-docker

project:
  name: my-service
  groupId: com.example.team
  repository: Nexus_PROD
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
	if cfg.Project.GroupID != "com.example.team" {
		t.Errorf("project.groupId = %q, want com.example.team", cfg.Project.GroupID)
	}
	if cfg.Project.Repository != "Nexus_PROD" {
		t.Errorf("project.repository = %q, want Nexus_PROD", cfg.Project.Repository)
	}
	if cfg.Coin.Template != "python-uv-app" {
		t.Errorf("coin.template = %q, want python-uv-app", cfg.Coin.Template)
	}
	if cfg.Jenkins.Runtime["python"] != "3.13" {
		t.Errorf("jenkins.runtime.python = %q, want 3.13", cfg.Jenkins.Runtime["python"])
	}
	if cfg.Jenkins.Credentials.Docker != "nexus-docker" {
		t.Errorf("jenkins.credentials.docker = %q, want nexus-docker", cfg.Jenkins.Credentials.Docker)
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

func TestLoad_MissingProjectName(t *testing.T) {
	path := writeConfig(t, `
coin:
  template: go-app
jenkins:
  credentials:
    docker: nexus-docker
project:
  name: ""
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing project.name")
	}
}

func TestLoad_MissingTemplate(t *testing.T) {
	path := writeConfig(t, `
jenkins:
  credentials:
    docker: nexus-docker
project:
  name: my-service
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing coin.template")
	}
}

func TestLoad_MissingDockerCredential(t *testing.T) {
	path := writeConfig(t, `
coin:
  template: go-app
project:
  name: svc
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing jenkins.credentials.docker")
	}
}

func TestStageIsEnabled(t *testing.T) {
	trueVal := true
	falseVal := false

	cases := []struct {
		s    Stage
		want bool
	}{
		{Stage{Enabled: nil}, true},
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
		Jenkins: JenkinsConfig{
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
