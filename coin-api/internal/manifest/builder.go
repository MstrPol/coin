package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Builder struct{}

type BuildOptions struct{}

type GPRelease struct {
	Name         string
	Version      string
	Destinations Destinations
	Parts        Composition
	Content      ContentBundle
	Branching    BranchingBundle
}

type Destinations struct {
	ImageRegistryPrefix    string `json:"imageRegistryPrefix"`
	BuildCacheEnabled      bool   `json:"buildCacheEnabled"`
	ArtifactRepositoryBase string `json:"artifactRepositoryBase"`
}

type Composition struct {
	AgentImage            string
	AgentDigest           string
	GPContentName         string
	GPContentVersion      string
	PipelineVersion       string
	BranchingModelName    string
	BranchingModelVersion string
}

type ContentBundle struct {
	Name                  string
	Version               string
	BundleURL             string
	BundleSHA256          string
	BuildControls         map[string]any
	Capabilities          map[string]any
	Parameters            []Parameter
	SchemaArtifactKey     string
	SchemaSHA256          string
	ContainerfileKey      string
	ContainerfileSHA256   string
	Containerfiles        []NamedContentRef
	Targets               []BuildTarget
	Deliverables          []Deliverable
	BuildEngine           string
	BuildkitTargets       map[string]string
	DockerfilePath        string
	DockerfileImageTarget string
	DockerfileTestTarget  string
	Stages                []TypedStage
}

type NamedContentRef struct {
	ID          string `json:"id"`
	ArtifactKey string `json:"artifactKey,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
}

type Parameter struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Default       any      `json:"default,omitempty"`
	Required      bool     `json:"required,omitempty"`
	AllowedValues []string `json:"allowedValues,omitempty"`
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
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	TargetID string               `json:"targetId"`
	Image    *ImageDeliverable    `json:"image,omitempty"`
	Artifact *ArtifactDeliverable `json:"artifact,omitempty"`
}

type ImageDeliverable struct {
	RepositorySuffix string `json:"repositorySuffix,omitempty"`
}

type ArtifactDeliverable struct {
	Format string   `json:"format,omitempty"`
	Paths  []string `json:"paths,omitempty"`
}

type TypedStage struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	When  string      `json:"when,omitempty"`
	Steps []StageStep `json:"steps,omitempty"`
}

type StageStep struct {
	Action        string            `json:"action"`
	TargetID      string            `json:"targetId,omitempty"`
	DeliverableID string            `json:"deliverableId,omitempty"`
	Run           *InlineRunStep    `json:"run,omitempty"`
	Build         *InlineBuildStep  `json:"build,omitempty"`
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
	Image         *ImageDeliverable        `json:"image,omitempty"`
	Artifact      *ArtifactDeliverable     `json:"artifact,omitempty"`
}

type InlinePublishStep struct {
	BuildStepID string `json:"buildStepId"`
}

func (c ContentBundle) IsInlinePipeline() bool {
	for _, stage := range c.Stages {
		for _, step := range stage.Steps {
			switch step.Action {
			case "run", "build", "publish":
				return true
			}
		}
	}
	return false
}

func (b Builder) Build(release GPRelease, opts BuildOptions) (map[string]any, string, error) {
	if err := validateDestinations(release.Destinations); err != nil {
		return nil, "", err
	}
	resolvedDockerfile := ".coin/Containerfile"

	doc := map[string]any{
		"manifestVersion": 1,
		"goldenPath": map[string]string{
			"name":    release.Name,
			"version": release.Version,
		},
		"runtime": map[string]string{
			"image":  release.Parts.AgentImage,
			"digest": release.Parts.AgentDigest,
		},
		"destinations": map[string]any{
			"imageRegistryPrefix":    release.Destinations.ImageRegistryPrefix,
			"buildCacheEnabled":      release.Destinations.BuildCacheEnabled,
			"artifactRepositoryBase": release.Destinations.ArtifactRepositoryBase,
		},
		"pipeline": map[string]any{
			"stages": b.stages(release),
		},
		"validateSchema": contentRef(release.Name, release.Version, release.Content.SchemaArtifactKey, release.Content.SchemaSHA256),
	}
	if !release.Content.IsInlinePipeline() {
		doc["build"] = b.buildSection(release, opts, resolvedDockerfile)
	}
	if len(release.Content.Parameters) > 0 {
		doc["parameters"] = release.Content.Parameters
	}
	if len(release.Content.Deliverables) > 0 && !release.Content.IsInlinePipeline() {
		doc["deliverables"] = release.Content.Deliverables
	}
	if len(release.Content.Containerfiles) > 0 && !release.Content.IsInlinePipeline() {
		doc["artifacts"] = map[string]any{
			"containerfiles": b.containerfileRefs(release),
		}
	}
	if len(release.Content.Capabilities) > 0 {
		doc["capabilities"] = release.Content.Capabilities
	}
	if branchingDoc := b.branchingSection(release.Branching); branchingDoc != nil {
		doc["branching"] = branchingDoc
	}

	raw, err := canonicalJSON(doc)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(raw)
	hash := "sha256:" + hex.EncodeToString(sum[:])
	doc["manifestHash"] = hash
	return doc, hash, nil
}

func validateDestinations(dest Destinations) error {
	if strings.TrimSpace(dest.ImageRegistryPrefix) == "" {
		return fmt.Errorf("gp release destinations.imageRegistryPrefix is required")
	}
	if strings.TrimSpace(dest.ArtifactRepositoryBase) == "" {
		return fmt.Errorf("gp release destinations.artifactRepositoryBase is required")
	}
	return nil
}

func (b Builder) branchingSection(bundle BranchingBundle) map[string]any {
	version := strings.TrimSpace(bundle.Version)
	if version == "" {
		return nil
	}
	name := strings.TrimSpace(bundle.Name)
	if name == "" {
		if n, ok := bundle.Rules["name"].(string); ok {
			name = strings.TrimSpace(n)
		}
	}
	out := map[string]any{"version": version}
	if name != "" {
		out["name"] = name
	}
	for k, v := range bundle.Rules {
		if k == "name" || k == "version" {
			continue
		}
		out[k] = v
	}
	return out
}

func (b Builder) buildSection(release GPRelease, opts BuildOptions, resolvedDockerfile string) map[string]any {
	if len(release.Content.Targets) > 0 {
		return map[string]any{
			"targets": release.Content.Targets,
		}
	}
	engine := release.Content.BuildEngine
	if engine == "" {
		engine = "buildkit"
	}
	out := map[string]any{
		"engine": engine,
	}
	if engine != "buildkit" {
		switch engine {
		case "dockerfile":
			dockerfilePath := strings.TrimSpace(release.Content.DockerfilePath)
			if dockerfilePath == "" {
				dockerfilePath = "Dockerfile"
			}
			df := map[string]any{
				"dockerfile": dockerfilePath,
			}
			if target := strings.TrimSpace(release.Content.DockerfileImageTarget); target != "" {
				df["imageTarget"] = target
			}
			if target := strings.TrimSpace(release.Content.DockerfileTestTarget); target != "" {
				df["testTarget"] = target
			}
			out["dockerfile"] = df
		}
		return out
	}
	targets := release.Content.BuildkitTargets
	if len(targets) == 0 {
		targets = map[string]string{
			"validate": "validate",
			"test":     "test",
			"image":    "runtime",
			"artifact": "artifact",
		}
	}
	buildkit := map[string]any{
		"dockerfile": resolvedDockerfile,
		"targets":    targets,
		"containerfile": contentRef(
			release.Content.contentName(release.Name),
			release.Content.contentVersion(release.Version),
			release.Content.ContainerfileKey,
			release.Content.ContainerfileSHA256,
		),
	}
	out["buildkit"] = buildkit
	return out
}

func (b Builder) stages(release GPRelease) []map[string]any {
	out := make([]map[string]any, 0, len(release.Content.Stages))
	for _, stage := range release.Content.Stages {
		item := map[string]any{
			"id":   stage.ID,
			"name": stage.Name,
		}
		if stage.When != "" {
			item["when"] = stage.When
		}
		if len(stage.Steps) > 0 {
			item["steps"] = stage.Steps
		}
		out = append(out, item)
	}
	return out
}

func (b Builder) containerfileRefs(release GPRelease) []map[string]string {
	out := make([]map[string]string, 0, len(release.Content.Containerfiles))
	for _, cf := range release.Content.Containerfiles {
		ref := contentRef(release.Content.contentName(release.Name), release.Content.contentVersion(release.Version), cf.ArtifactKey, cf.SHA256)
		ref["id"] = cf.ID
		out = append(out, ref)
	}
	return out
}

func (c ContentBundle) contentName(fallback string) string {
	if strings.TrimSpace(c.Name) != "" {
		return c.Name
	}
	return fallback
}

func (c ContentBundle) contentVersion(fallback string) string {
	if strings.TrimSpace(c.Version) != "" {
		return c.Version
	}
	return fallback
}

func contentRef(gpName, gpVersion, artifactKey, sha256sum string) map[string]string {
	return map[string]string{
		"url":    ContentArtifactURL(gpName, gpVersion, artifactKey),
		"sha256": sha256sum,
	}
}

func canonicalJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}
	var normalized any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		return nil, err
	}
	sorted := sortKeys(normalized)
	return json.Marshal(sorted)
}

func sortKeys(v any) any {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]any, len(t))
		for _, k := range keys {
			out[k] = sortKeys(t[k])
		}
		return out
	case []any:
		for i, item := range t {
			t[i] = sortKeys(item)
		}
		return t
	default:
		return v
	}
}
