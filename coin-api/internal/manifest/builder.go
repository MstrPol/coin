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

type BuildOptions struct {
	Project      string
	RegistryHost string
}

type GPRelease struct {
	Name    string
	Version string
	Parts   Composition
	Content ContentBundle
}

type Composition struct {
	ExecutorVersion  string
	AgentImage       string
	AgentDigest      string
	ExecutorURL      string
	ExecutorSHA256   string
	LibName          string
	LibVersion       string
	GPContentName    string
	GPContentVersion string
	PipelineVersion  string
}

type ContentBundle struct {
	BundleURL            string
	BundleSHA256         string
	BuildControls        map[string]any
	Capabilities         map[string]any
	SchemaArtifactKey    string
	SchemaSHA256         string
	ContainerfileKey     string
	ContainerfileSHA256  string
	BuildEngine          string
	BuildkitDockerfile   string
	BuildkitTargets      map[string]string
	BuildpackBuilder        string
	BuildpackRunImage       string
	DockerfileImageTarget   string
	DockerfileTestTarget    string
	CacheRefTemplate        string
	Stages               []TypedStage
}

type TypedStage struct {
	ID   string
	Name string
	When string
}

func (b Builder) Build(release GPRelease, opts BuildOptions) (map[string]any, string, error) {
	resolvedDockerfile := ".coin/Containerfile"
	buildDoc := b.buildSection(release, opts, resolvedDockerfile)

	doc := map[string]any{
		"manifestVersion": 1,
		"goldenPath": map[string]string{
			"name":    release.Name,
			"version": release.Version,
		},
		"executor": map[string]string{
			"version": release.Parts.ExecutorVersion,
			"url":     RuntimeNexusURL(release.Parts.ExecutorURL),
			"sha256":  release.Parts.ExecutorSHA256,
		},
		"runtime": map[string]string{
			"image":  release.Parts.AgentImage,
			"digest": release.Parts.AgentDigest,
		},
		"build": buildDoc,
		"pipeline": map[string]any{
			"stages": b.stages(release),
		},
		"validateSchema": contentRef(release.Name, release.Version, release.Content.SchemaArtifactKey, release.Content.SchemaSHA256),
		"credentials": map[string]string{
			"docker": "nexus-docker",
		},
	}
	if len(release.Content.Capabilities) > 0 {
		doc["capabilities"] = release.Content.Capabilities
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

func (b Builder) buildSection(release GPRelease, opts BuildOptions, resolvedDockerfile string) map[string]any {
	engine := release.Content.BuildEngine
	if engine == "" {
		engine = "buildkit"
	}
	out := map[string]any{
		"engine": engine,
	}
	if engine != "buildkit" {
		switch engine {
		case "buildpack":
			bp := map[string]any{
				"builder": release.Content.BuildpackBuilder,
			}
			if runImage := strings.TrimSpace(release.Content.BuildpackRunImage); runImage != "" {
				bp["runImage"] = runImage
			}
			if cacheRef := resolveCacheRef(release.Content.CacheRefTemplate, opts); cacheRef != "" {
				bp["cacheRef"] = cacheRef
			}
			out["buildpack"] = bp
		case "dockerfile":
			df := map[string]any{
				"dockerfile": resolvedDockerfile,
				"containerfile": contentRef(
					release.Name,
					release.Version,
					release.Content.ContainerfileKey,
					release.Content.ContainerfileSHA256,
				),
			}
			if target := strings.TrimSpace(release.Content.DockerfileImageTarget); target != "" {
				df["imageTarget"] = target
			}
			if target := strings.TrimSpace(release.Content.DockerfileTestTarget); target != "" {
				df["testTarget"] = target
			}
			if cacheRef := resolveCacheRef(release.Content.CacheRefTemplate, opts); cacheRef != "" {
				df["cacheRef"] = cacheRef
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
	cacheRef := resolveCacheRef(release.Content.CacheRefTemplate, opts)
	buildkit := map[string]any{
		"dockerfile": resolvedDockerfile,
		"targets":    targets,
		"containerfile": contentRef(
			release.Name,
			release.Version,
			release.Content.ContainerfileKey,
			release.Content.ContainerfileSHA256,
		),
	}
	if cacheRef != "" {
		buildkit["cacheRef"] = cacheRef
	}
	out["buildkit"] = buildkit
	return out
}

func resolveCacheRef(template string, opts BuildOptions) string {
	if strings.TrimSpace(template) == "" {
		return ""
	}
	host := strings.TrimSpace(opts.RegistryHost)
	if host == "" {
		host = "localhost:8082"
	}
	project := strings.TrimSpace(opts.Project)
	if project == "" {
		project = "app"
	}
	out := template
	out = strings.ReplaceAll(out, "{{registryHost}}", host)
	out = strings.ReplaceAll(out, "{{registry}}", host)
	out = strings.ReplaceAll(out, "{{project}}", project)
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
		out = append(out, item)
	}
	return out
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
