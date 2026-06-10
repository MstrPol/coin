package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

type Builder struct{}

type GPRelease struct {
	Name    string
	Version string
	Parts   Composition
	Content ContentBundle
}

type Composition struct {
	ExecutorVersion   string
	AgentImage        string
	AgentDigest       string
	ExecutorURL       string
	ExecutorSHA256    string
	PipelineVersion   string
	ValidateVersion   string
	DockerfileVersion string
}

type ContentBundle struct {
	SchemaArtifactKey         string
	SchemaSHA256              string
	DockerfileArtifactKey     string
	DockerfileSHA256          string
	OrchestrationArtifactKey  string
	OrchestrationSHA256       string
	Stages                    []StageScript
}

type StageScript struct {
	Name        string
	When        string
	ArtifactKey string
	SHA256      string
}

func (b Builder) Build(release GPRelease) (map[string]any, string, error) {
	doc := map[string]any{
		"manifestVersion": 1,
		"goldenPath": map[string]string{
			"name":    release.Name,
			"version": release.Version,
		},
		"executor": map[string]string{
			"version": release.Parts.ExecutorVersion,
			"url":     release.Parts.ExecutorURL,
			"sha256":  release.Parts.ExecutorSHA256,
		},
		"runtime": map[string]string{
			"image":  release.Parts.AgentImage,
			"digest": release.Parts.AgentDigest,
		},
		"pipeline": map[string]any{
			"stages": b.stages(release),
		},
		"orchestration": contentRef(release.Name, release.Version, release.Content.OrchestrationArtifactKey, release.Content.OrchestrationSHA256),
		"validateSchema": contentRef(release.Name, release.Version, release.Content.SchemaArtifactKey, release.Content.SchemaSHA256),
		"dockerfileTemplate": contentRef(release.Name, release.Version, release.Content.DockerfileArtifactKey, release.Content.DockerfileSHA256),
		"credentials": map[string]string{
			"docker": "nexus-docker",
		},
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

func (b Builder) stages(release GPRelease) []map[string]any {
	out := make([]map[string]any, 0, len(release.Content.Stages))
	for _, stage := range release.Content.Stages {
		item := map[string]any{
			"name": stage.Name,
			"script": contentRef(release.Name, release.Version, stage.ArtifactKey, stage.SHA256),
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
