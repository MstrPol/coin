package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"coin.local/coin-executor/internal/pin"
)

type Manifest struct {
	ManifestVersion int          `json:"manifestVersion"`
	ManifestHash    string       `json:"manifestHash"`
	GoldenPath      GoldenPath   `json:"goldenPath"`
	Executor        Executor     `json:"executor"`
	Runtime         Runtime      `json:"runtime"`
	Build           Build        `json:"build"`
	Pipeline        Pipeline     `json:"pipeline"`
	ValidateSchema  ContentRef   `json:"validateSchema"`
	Credentials     Credentials  `json:"credentials"`
	Capabilities    Capabilities `json:"capabilities"`
	Branching       *Branching   `json:"branching,omitempty"`
}

type Branching struct {
	Name     string        `json:"name"`
	Version  string        `json:"version"`
	Branches []BranchRule  `json:"branches"`
}

type BranchRule struct {
	Name       string            `json:"name"`
	Pattern    string            `json:"pattern"`
	Versioning BranchVersioning  `json:"versioning"`
	Publish    bool              `json:"publish"`
}

type BranchVersioning struct {
	Template string `json:"template"`
}

type Capabilities struct {
	Deliverables []string `json:"deliverables"`
}

func (m *Manifest) AllowedDeliverableTypes() []string {
	if len(m.Capabilities.Deliverables) > 0 {
		return m.Capabilities.Deliverables
	}
	return nil
}

type GoldenPath struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Executor struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	SHA256  string `json:"sha256"`
}

type Runtime struct {
	Image  string `json:"image"`
	Digest string `json:"digest"`
}

type Build struct {
	Engine     string                  `json:"engine"`
	Buildkit   *BuildkitConfig         `json:"buildkit,omitempty"`
	Dockerfile *DockerfileEngineConfig `json:"dockerfile,omitempty"`
}

type BuildkitConfig struct {
	Dockerfile    string            `json:"dockerfile"`
	Targets       map[string]string `json:"targets"`
	CacheRef      string            `json:"cacheRef,omitempty"`
	Containerfile ContentRef        `json:"containerfile"`
}

// DockerfileEngineConfig is a simplified BuildKit dockerfile.v0 policy (no targets map).
type DockerfileEngineConfig struct {
	Dockerfile    string     `json:"dockerfile"`
	ImageTarget   string     `json:"imageTarget,omitempty"`
	TestTarget    string     `json:"testTarget,omitempty"`
	CacheRef      string     `json:"cacheRef,omitempty"`
	Containerfile ContentRef `json:"containerfile"`
}

type Pipeline struct {
	Stages []Stage `json:"stages"`
}

type Stage struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	When string `json:"when,omitempty"`
}

func (s Stage) Key() string {
	if strings.TrimSpace(s.ID) != "" {
		return strings.TrimSpace(s.ID)
	}
	return strings.ToLower(strings.TrimSpace(s.Name))
}

type ContentRef struct {
	URL    string `json:"url"`
	GitRef string `json:"gitRef"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Credentials struct {
	Docker string `json:"docker"`
}

func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if m.GoldenPath.Name == "" || m.GoldenPath.Version == "" {
		return nil, fmt.Errorf("manifest goldenPath name/version required")
	}
	return &m, nil
}

func (m *Manifest) MatchesConfig(goldenPath, configPin string) error {
	if m.GoldenPath.Name != goldenPath {
		return fmt.Errorf("manifest gp %s != config %s", m.GoldenPath.Name, goldenPath)
	}
	p, err := pin.Parse(configPin)
	if err != nil {
		return fmt.Errorf("config pin: %w", err)
	}
	if !p.Satisfies(m.GoldenPath.Version) {
		return fmt.Errorf("manifest gp %s@%s does not satisfy config pin %q",
			m.GoldenPath.Name, m.GoldenPath.Version, configPin)
	}
	return nil
}

func (m *Manifest) URLRefsOnly() bool {
	refs := []ContentRef{m.ValidateSchema}
	if m.Build.Buildkit != nil {
		refs = append(refs, m.Build.Buildkit.Containerfile)
	}
	if m.Build.Dockerfile != nil {
		refs = append(refs, m.Build.Dockerfile.Containerfile)
	}
	for _, ref := range refs {
		if strings.TrimSpace(ref.URL) == "" {
			return false
		}
	}
	return len(refs) > 0
}

func (m *Manifest) ContentGitRefs() []string {
	seen := map[string]struct{}{}
	add := func(ref string) {
		if ref != "" {
			seen[ref] = struct{}{}
		}
	}
	add(m.ValidateSchema.GitRef)
	if m.Build.Buildkit != nil {
		add(m.Build.Buildkit.Containerfile.GitRef)
	}
	if m.Build.Dockerfile != nil {
		add(m.Build.Dockerfile.Containerfile.GitRef)
	}
	out := make([]string, 0, len(seen))
	for ref := range seen {
		out = append(out, ref)
	}
	return out
}

func (m *Manifest) BuildkitTarget(key string) string {
	if m.Build.Buildkit == nil {
		return ""
	}
	if target := m.Build.Buildkit.Targets[key]; target != "" {
		return target
	}
	return key
}
