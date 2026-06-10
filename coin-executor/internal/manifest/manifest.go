package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"coin.local/coin-executor/internal/pin"
)

type Manifest struct {
	ManifestVersion    int               `json:"manifestVersion"`
	ManifestHash       string            `json:"manifestHash"`
	GoldenPath         GoldenPath        `json:"goldenPath"`
	Executor           Executor          `json:"executor"`
	Runtime            Runtime           `json:"runtime"`
	Pipeline           Pipeline          `json:"pipeline"`
	ValidateSchema     ContentRef        `json:"validateSchema"`
	DockerfileTemplate ContentRef        `json:"dockerfileTemplate"`
	Credentials        Credentials       `json:"credentials"`
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

type Pipeline struct {
	Stages []Stage `json:"stages"`
}

type Stage struct {
	Name   string     `json:"name"`
	When   string     `json:"when,omitempty"`
	Script ContentRef `json:"script"`
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
	refs := []ContentRef{m.ValidateSchema, m.DockerfileTemplate}
	for _, stage := range m.Pipeline.Stages {
		refs = append(refs, stage.Script)
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
	add(m.DockerfileTemplate.GitRef)
	for _, stage := range m.Pipeline.Stages {
		add(stage.Script.GitRef)
	}
	out := make([]string, 0, len(seen))
	for ref := range seen {
		out = append(out, ref)
	}
	return out
}
