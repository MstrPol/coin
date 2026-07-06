package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"coin.local/coin-executor/internal/deliverables"
	"coin.local/coin-executor/internal/pin"
)

type Manifest struct {
	ManifestVersion int           `json:"manifestVersion"`
	ManifestHash    string        `json:"manifestHash"`
	GoldenPath      GoldenPath    `json:"goldenPath"`
	Runtime         Runtime       `json:"runtime"`
	Destinations    Destinations  `json:"destinations"`
	Parameters      []Parameter   `json:"parameters,omitempty"`
	Build           Build         `json:"build"`
	Deliverables    []Deliverable `json:"deliverables,omitempty"`
	Artifacts       Artifacts     `json:"artifacts,omitempty"`
	Pipeline        Pipeline      `json:"pipeline"`
	ValidateSchema  ContentRef    `json:"validateSchema"`
	Capabilities    Capabilities  `json:"capabilities"`
	Branching       *Branching    `json:"branching,omitempty"`
}

type Branching struct {
	Name     string       `json:"name"`
	Version  string       `json:"version"`
	Branches []BranchRule `json:"branches"`
}

type BranchRule struct {
	Name       string           `json:"name"`
	Pattern    string           `json:"pattern"`
	Versioning BranchVersioning `json:"versioning"`
	Publish    bool             `json:"publish"`
}

type BranchVersioning struct {
	Template string `json:"template"`
}

type Capabilities struct {
	Deliverables []string `json:"deliverables"`
}

func (m *Manifest) AllowedDeliverableTypes() []string {
	if len(m.Deliverables) > 0 {
		out := make([]string, 0, len(m.Deliverables))
		seen := map[string]struct{}{}
		for _, item := range m.Deliverables {
			if _, ok := seen[item.Type]; ok {
				continue
			}
			seen[item.Type] = struct{}{}
			out = append(out, item.Type)
		}
		return out
	}
	if len(m.Capabilities.Deliverables) > 0 {
		return m.Capabilities.Deliverables
	}
	return nil
}

func (m *Manifest) DeliverableSpecs() map[string]deliverables.Spec {
	if m.IsInlinePipeline() {
		out := make(map[string]deliverables.Spec)
		for _, stage := range m.Pipeline.Stages {
			for _, step := range stage.Steps {
				if step.Action != "build" || step.Build == nil {
					continue
				}
				item := step.Build
				spec := deliverables.Spec{Type: item.Type}
				switch item.Type {
				case "image":
					spec.Context = "."
				case "liquibase-image":
					spec.Path = "liquibase"
				case "artifact":
					spec.Format = item.Artifact.Format
					if spec.Format == "" {
						spec.Format = "zip"
					}
					if len(item.Artifact.Paths) > 0 {
						spec.Source = item.Artifact.Paths[0]
					} else {
						spec.Source = ".coin/artifacts/" + item.ID + ".zip"
					}
				}
				out[item.ID] = spec
			}
		}
		return out
	}
	if len(m.Deliverables) > 0 {
		out := make(map[string]deliverables.Spec, len(m.Deliverables))
		for _, item := range m.Deliverables {
			spec := deliverables.Spec{Type: item.Type}
			switch item.Type {
			case "image":
				spec.Context = "."
			case "liquibase-image":
				spec.Path = "liquibase"
			case "artifact":
				spec.Format = item.Artifact.Format
				if spec.Format == "" {
					spec.Format = "zip"
				}
				if len(item.Artifact.Paths) > 0 {
					spec.Source = item.Artifact.Paths[0]
				} else {
					spec.Source = ".coin/artifacts/" + item.ID + ".zip"
				}
			}
			out[item.ID] = spec
		}
		return out
	}
	out := make(map[string]deliverables.Spec, len(m.Capabilities.Deliverables))
	for _, typ := range m.Capabilities.Deliverables {
		switch typ {
		case "image":
			out["app"] = deliverables.Spec{Type: "image", Context: "."}
		case "liquibase-image":
			out["liquibase"] = deliverables.Spec{Type: "liquibase-image", Path: "liquibase"}
		case "artifact":
			out["artifact"] = deliverables.Spec{Type: "artifact", Format: "zip", Source: ".coin/artifacts/app.zip"}
		default:
			out[typ] = deliverables.Spec{Type: typ}
		}
	}
	return out
}

func (m *Manifest) ValidateDeliverables() error {
	if m.IsInlinePipeline() {
		count := 0
		for _, stage := range m.Pipeline.Stages {
			for _, step := range stage.Steps {
				if step.Action == "build" && step.Build != nil && strings.TrimSpace(step.Build.ID) != "" {
					count++
				}
			}
		}
		if count == 0 {
			return fmt.Errorf("inline pipeline requires at least one build step")
		}
		return nil
	}
	if len(m.Deliverables) > 0 {
		allowed := map[string]struct{}{}
		for _, typ := range deliverables.P0Types {
			allowed[typ] = struct{}{}
		}
		seen := map[string]struct{}{}
		for _, item := range m.Deliverables {
			if strings.TrimSpace(item.ID) == "" {
				return fmt.Errorf("manifest deliverable id is required")
			}
			if _, dup := seen[item.ID]; dup {
				return fmt.Errorf("manifest declares duplicate deliverable %q", item.ID)
			}
			seen[item.ID] = struct{}{}
			if _, ok := allowed[item.Type]; !ok {
				return fmt.Errorf("manifest deliverable type %q is not supported in P0", item.Type)
			}
			if strings.TrimSpace(item.TargetID) == "" {
				return fmt.Errorf("manifest deliverable %q targetId is required", item.ID)
			}
		}
		return nil
	}
	if len(m.Capabilities.Deliverables) == 0 {
		return fmt.Errorf("manifest capabilities.deliverables is required")
	}
	allowed := map[string]struct{}{}
	for _, typ := range deliverables.P0Types {
		allowed[typ] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, typ := range m.Capabilities.Deliverables {
		if _, ok := allowed[typ]; !ok {
			return fmt.Errorf("manifest deliverable type %q is not supported in P0", typ)
		}
		if _, dup := seen[typ]; dup {
			return fmt.Errorf("manifest declares multiple %q deliverables", typ)
		}
		seen[typ] = struct{}{}
	}
	return nil
}

func (m *Manifest) HasDeliverable(typ string) bool {
	for _, item := range m.Deliverables {
		if item.Type == typ {
			return true
		}
	}
	for _, item := range m.Capabilities.Deliverables {
		if item == typ {
			return true
		}
	}
	return false
}

type GoldenPath struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Runtime struct {
	Image  string `json:"image"`
	Digest string `json:"digest"`
}

type Destinations struct {
	ImageRegistryPrefix    string `json:"imageRegistryPrefix"`
	BuildCacheEnabled      bool   `json:"buildCacheEnabled"`
	ArtifactRepositoryBase string `json:"artifactRepositoryBase"`
}

type Parameter struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Default       any      `json:"default,omitempty"`
	Required      bool     `json:"required,omitempty"`
	AllowedValues []string `json:"allowedValues,omitempty"`
}

type Build struct {
	Engine     string                  `json:"engine"`
	Targets    []BuildTarget           `json:"targets,omitempty"`
	Buildkit   *BuildkitConfig         `json:"buildkit,omitempty"`
	Dockerfile *DockerfileEngineConfig `json:"dockerfile,omitempty"`
}

type BuildTarget struct {
	ID            string `json:"id"`
	Engine        string `json:"engine"`
	Containerfile string `json:"containerfile,omitempty"`
	Target        string `json:"target,omitempty"`
	Dockerfile    string `json:"dockerfile,omitempty"`
	Output        string `json:"output,omitempty"`
}

type Deliverable struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	TargetID string              `json:"targetId"`
	Image    ImageDeliverable    `json:"image,omitempty"`
	Artifact ArtifactDeliverable `json:"artifact,omitempty"`
}

type ImageDeliverable struct {
	RepositorySuffix string `json:"repositorySuffix,omitempty"`
}

type ArtifactDeliverable struct {
	Format string   `json:"format,omitempty"`
	Paths  []string `json:"paths,omitempty"`
}

type Artifacts struct {
	Containerfiles []NamedContentRef `json:"containerfiles,omitempty"`
}

type NamedContentRef struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type BuildkitConfig struct {
	Dockerfile    string            `json:"dockerfile"`
	Targets       map[string]string `json:"targets"`
	Containerfile ContentRef        `json:"containerfile"`
}

// DockerfileEngineConfig is a simplified BuildKit dockerfile.v0 policy (no targets map).
type DockerfileEngineConfig struct {
	Dockerfile    string     `json:"dockerfile"`
	ImageTarget   string     `json:"imageTarget,omitempty"`
	TestTarget    string     `json:"testTarget,omitempty"`
	Containerfile ContentRef `json:"containerfile"`
}

type Pipeline struct {
	Stages []Stage `json:"stages"`
}

type Stage struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	When  string      `json:"when,omitempty"`
	Steps []StageStep `json:"steps,omitempty"`
}

type StageStep struct {
	Action        string             `json:"action"`
	TargetID      string             `json:"targetId,omitempty"`
	DeliverableID string             `json:"deliverableId,omitempty"`
	Run           *InlineRunStep     `json:"run,omitempty"`
	Build         *InlineBuildStep   `json:"build,omitempty"`
	Publish       *InlinePublishStep `json:"publish,omitempty"`
}

type InlineContainerfileStep struct {
	Body       string            `json:"body,omitempty"`
	ContentRef map[string]string `json:"contentRef,omitempty"`
	Digest     string            `json:"digest,omitempty"`
}

type InlineDockerfileStep struct {
	Path   string `json:"path"`
	Target string `json:"target,omitempty"`
}

type InlineRunStep struct {
	Engine        string                   `json:"engine"`
	Target        string                   `json:"target,omitempty"`
	Output        string                   `json:"output,omitempty"`
	Containerfile *InlineContainerfileStep `json:"containerfile,omitempty"`
	Dockerfile    *InlineDockerfileStep    `json:"dockerfile,omitempty"`
}

type InlineBuildStep struct {
	ID            string                   `json:"id"`
	Type          string                   `json:"type"`
	Engine        string                   `json:"engine"`
	Target        string                   `json:"target,omitempty"`
	Containerfile *InlineContainerfileStep `json:"containerfile,omitempty"`
	Dockerfile    *InlineDockerfileStep    `json:"dockerfile,omitempty"`
	Image         ImageDeliverable         `json:"image,omitempty"`
	Artifact      ArtifactDeliverable      `json:"artifact,omitempty"`
}

type InlinePublishStep struct {
	BuildStepID string `json:"buildStepId"`
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
	if strings.TrimSpace(m.Destinations.ImageRegistryPrefix) == "" {
		return nil, fmt.Errorf("manifest destinations.imageRegistryPrefix required")
	}
	if strings.TrimSpace(m.Destinations.ArtifactRepositoryBase) == "" {
		return nil, fmt.Errorf("manifest destinations.artifactRepositoryBase required")
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
	for _, ref := range m.Artifacts.Containerfiles {
		refs = append(refs, ContentRef{URL: ref.URL, SHA256: ref.SHA256})
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

func (m *Manifest) IsVNext() bool {
	if m.IsInlinePipeline() {
		return false
	}
	return len(m.Build.Targets) > 0 || len(m.Deliverables) > 0
}

func (m *Manifest) IsInlinePipeline() bool {
	for _, stage := range m.Pipeline.Stages {
		for _, step := range stage.Steps {
			switch step.Action {
			case "run", "build", "publish":
				return true
			}
		}
	}
	return false
}

func (m *Manifest) InlineBuild(id string) (InlineBuildStep, bool) {
	for _, stage := range m.Pipeline.Stages {
		for _, step := range stage.Steps {
			if step.Action == "build" && step.Build != nil && step.Build.ID == id {
				return *step.Build, true
			}
		}
	}
	return InlineBuildStep{}, false
}

func (m *Manifest) Target(id string) (BuildTarget, bool) {
	for _, target := range m.Build.Targets {
		if target.ID == id {
			return target, true
		}
	}
	return BuildTarget{}, false
}

func (m *Manifest) Deliverable(id string) (Deliverable, bool) {
	for _, item := range m.Deliverables {
		if item.ID == id {
			return item, true
		}
	}
	return Deliverable{}, false
}

func (m *Manifest) Containerfile(id string) (NamedContentRef, bool) {
	for _, item := range m.Artifacts.Containerfiles {
		if item.ID == id {
			return item, true
		}
	}
	return NamedContentRef{}, false
}

func (m *Manifest) ResolvedParameters() map[string]any {
	out := make(map[string]any, len(m.Parameters))
	for _, item := range m.Parameters {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		out[item.Name] = item.Default
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
