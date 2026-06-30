package gpcontent

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const SchemaVersion = 2

type Doc struct {
	SchemaVersion int          `yaml:"schemaVersion"`
	Name          string       `yaml:"name"`
	Kind          string       `yaml:"kind"`
	Capabilities  Capabilities `yaml:"capabilities"`
	Build         Build        `yaml:"build"`
	Pipeline      Pipeline     `yaml:"pipeline"`
	Artifacts     Artifacts    `yaml:"artifacts"`
}

type Capabilities struct {
	Deliverables []string `yaml:"deliverables"`
}

type Build struct {
	Engine    string          `yaml:"engine"`
	Buildkit  *BuildkitBlock  `yaml:"buildkit,omitempty"`
	Dockerfile *DockerfileBlock `yaml:"dockerfile,omitempty"`
}

type BuildkitBlock struct {
	Targets          map[string]string `yaml:"targets"`
	CacheRefTemplate string            `yaml:"cacheRefTemplate"`
}

type DockerfileBlock struct {
	Path             string `yaml:"path"`
	ImageTarget      string `yaml:"imageTarget,omitempty"`
	TestTarget       string `yaml:"testTarget,omitempty"`
	CacheRefTemplate string `yaml:"cacheRefTemplate"`
}

type Pipeline struct {
	Stages []Stage `yaml:"stages"`
}

type Stage struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type Artifacts struct {
	ValidateSchema string `yaml:"validateSchema"`
	Containerfile  string `yaml:"containerfile,omitempty"`
}

type Issue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type PreviewResult struct {
	Valid    bool           `json:"valid"`
	Issues   []Issue        `json:"issues"`
	Warnings []string       `json:"warnings,omitempty"`
	Build    map[string]any `json:"build,omitempty"`
	Pipeline map[string]any `json:"pipeline,omitempty"`
	Capabilities map[string]any `json:"capabilities,omitempty"`
}

type PreviewOptions struct {
	ComponentName string
	Project       string
	RegistryHost  string
	HasContainerfileArtifact bool
}

func ParseDoc(raw []byte) (Doc, error) {
	var doc Doc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return Doc{}, err
	}
	return doc, nil
}

func ValidateDoc(doc Doc, opts PreviewOptions) (issues []Issue, warnings []string) {
	name := strings.TrimSpace(opts.ComponentName)
	if doc.SchemaVersion != SchemaVersion {
		issues = append(issues, Issue{Field: "content.yaml.schemaVersion", Message: "schemaVersion must be 2 (v1 is not supported)"})
	}
	if strings.TrimSpace(doc.Name) == "" {
		issues = append(issues, Issue{Field: "content.yaml.name", Message: "name is required"})
	} else if name != "" && doc.Name != name {
		issues = append(issues, Issue{
			Field:   "content.yaml.name",
			Message: fmt.Sprintf("name %q must match component name %q", doc.Name, name),
		})
	}
	if doc.Kind != "gp-content" {
		issues = append(issues, Issue{Field: "content.yaml.kind", Message: "kind must be gp-content"})
	}
	engine := strings.TrimSpace(doc.Build.Engine)
	switch engine {
	case "buildkit", "dockerfile":
	default:
		if engine == "buildpack" {
			issues = append(issues, Issue{Field: "content.yaml.build.engine", Message: "buildpack engine is not supported (hard cut)"})
		} else {
			issues = append(issues, Issue{Field: "content.yaml.build.engine", Message: "build.engine must be buildkit or dockerfile"})
		}
	}
	if len(doc.Capabilities.Deliverables) == 0 {
		issues = append(issues, Issue{Field: "content.yaml.capabilities.deliverables", Message: "at least one deliverable is required"})
	}
	hasArtifact := false
	for _, d := range doc.Capabilities.Deliverables {
		if d == "artifact" {
			hasArtifact = true
		}
	}
	switch engine {
	case "buildkit":
		if doc.Build.Buildkit == nil {
			issues = append(issues, Issue{Field: "content.yaml.build.buildkit", Message: "buildkit block is required"})
		} else {
			if len(doc.Build.Buildkit.Targets) == 0 {
				issues = append(issues, Issue{Field: "content.yaml.build.buildkit.targets", Message: "targets map is required"})
			}
			if strings.TrimSpace(doc.Build.Buildkit.CacheRefTemplate) == "" {
				issues = append(issues, Issue{Field: "content.yaml.build.buildkit.cacheRefTemplate", Message: "cacheRefTemplate is required"})
			}
		}
		if strings.TrimSpace(doc.Artifacts.Containerfile) == "" {
			issues = append(issues, Issue{Field: "content.yaml.artifacts.containerfile", Message: "containerfile artifact is required for buildkit engine"})
		}
		if !opts.HasContainerfileArtifact && strings.TrimSpace(doc.Artifacts.Containerfile) != "" {
			warnings = append(warnings, "containerfile artifact key set but dockerfiles/Containerfile body not in package")
		}
	case "dockerfile":
		if doc.Build.Dockerfile == nil {
			issues = append(issues, Issue{Field: "content.yaml.build.dockerfile", Message: "dockerfile block is required"})
		} else {
			if strings.TrimSpace(doc.Build.Dockerfile.Path) == "" {
				issues = append(issues, Issue{Field: "content.yaml.build.dockerfile.path", Message: "path is required"})
			}
			if strings.TrimSpace(doc.Build.Dockerfile.CacheRefTemplate) == "" {
				issues = append(issues, Issue{Field: "content.yaml.build.dockerfile.cacheRefTemplate", Message: "cacheRefTemplate is required"})
			}
		}
		if hasArtifact {
			issues = append(issues, Issue{Field: "content.yaml.capabilities.deliverables", Message: "artifact deliverable is not supported for dockerfile engine"})
		}
		if strings.TrimSpace(doc.Artifacts.Containerfile) != "" {
			issues = append(issues, Issue{Field: "content.yaml.artifacts.containerfile", Message: "containerfile artifact must not be set for BYO dockerfile engine"})
		}
		if doc.Build.Dockerfile != nil && strings.TrimSpace(doc.Build.Dockerfile.TestTarget) == "" {
			warnings = append(warnings, "testTarget not set; test stage may be skipped at runtime")
		}
	}
	if len(doc.Pipeline.Stages) == 0 {
		issues = append(issues, Issue{Field: "content.yaml.pipeline.stages", Message: "at least one pipeline stage is required"})
	} else {
		for i, st := range doc.Pipeline.Stages {
			if strings.TrimSpace(st.ID) == "" || strings.TrimSpace(st.Name) == "" {
				issues = append(issues, Issue{
					Field:   fmt.Sprintf("content.yaml.pipeline.stages[%d]", i),
					Message: "stage id and name are required",
				})
			}
		}
	}
	if strings.TrimSpace(doc.Artifacts.ValidateSchema) == "" {
		issues = append(issues, Issue{Field: "content.yaml.artifacts.validateSchema", Message: "validateSchema artifact is required"})
	}
	return issues, warnings
}

func Preview(doc Doc, opts PreviewOptions) PreviewResult {
	issues, warnings := ValidateDoc(doc, opts)
	out := PreviewResult{
		Valid:    len(issues) == 0,
		Issues:   issues,
		Warnings: warnings,
	}
	if len(doc.Capabilities.Deliverables) > 0 {
		out.Capabilities = map[string]any{
			"deliverables": doc.Capabilities.Deliverables,
		}
	}
	stages := make([]map[string]string, 0, len(doc.Pipeline.Stages))
	for _, st := range doc.Pipeline.Stages {
		stages = append(stages, map[string]string{"id": st.ID, "name": st.Name})
	}
	out.Pipeline = map[string]any{"stages": stages}

	engine := strings.TrimSpace(doc.Build.Engine)
	if engine == "" {
		engine = "buildkit"
	}
	buildDoc := map[string]any{"engine": engine}
	cacheRef := resolveCacheRef(cacheTemplate(doc), opts)
	switch engine {
	case "buildkit":
		if doc.Build.Buildkit != nil {
			bk := map[string]any{
				"dockerfile": ".coin/Containerfile",
				"targets":    doc.Build.Buildkit.Targets,
			}
			if cacheRef != "" {
				bk["cacheRef"] = cacheRef
			}
			if key := strings.TrimSpace(doc.Artifacts.Containerfile); key != "" {
				bk["containerfile"] = map[string]string{
					"artifactKey": key,
				}
			}
			buildDoc["buildkit"] = bk
		}
	case "dockerfile":
		if doc.Build.Dockerfile != nil {
			df := map[string]any{
				"dockerfile": doc.Build.Dockerfile.Path,
			}
			if t := strings.TrimSpace(doc.Build.Dockerfile.ImageTarget); t != "" {
				df["imageTarget"] = t
			}
			if t := strings.TrimSpace(doc.Build.Dockerfile.TestTarget); t != "" {
				df["testTarget"] = t
			}
			if cacheRef != "" {
				df["cacheRef"] = cacheRef
			}
			buildDoc["dockerfile"] = df
		}
	}
	out.Build = buildDoc
	return out
}

func cacheTemplate(doc Doc) string {
	switch strings.TrimSpace(doc.Build.Engine) {
	case "dockerfile":
		if doc.Build.Dockerfile != nil {
			return doc.Build.Dockerfile.CacheRefTemplate
		}
	case "buildkit":
		if doc.Build.Buildkit != nil {
			return doc.Build.Buildkit.CacheRefTemplate
		}
	}
	return ""
}

func resolveCacheRef(template string, opts PreviewOptions) string {
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

func ManifestSubset(doc Doc) map[string]any {
	subset := map[string]any{
		"capabilities": map[string]any{
			"deliverables": doc.Capabilities.Deliverables,
		},
		"build": map[string]any{
			"engine": doc.Build.Engine,
		},
		"pipeline": map[string]any{
			"stages": doc.Pipeline.Stages,
		},
		"validateSchema": map[string]string{
			"artifactKey": doc.Artifacts.ValidateSchema,
		},
	}
	if doc.Build.Buildkit != nil {
		subset["build"] = map[string]any{
			"engine": doc.Build.Engine,
			"buildkit": map[string]any{
				"targets":          doc.Build.Buildkit.Targets,
				"cacheRefTemplate": doc.Build.Buildkit.CacheRefTemplate,
			},
		}
	}
	if doc.Build.Dockerfile != nil {
		subset["build"] = map[string]any{
			"engine": doc.Build.Engine,
			"dockerfile": map[string]any{
				"path":             doc.Build.Dockerfile.Path,
				"imageTarget":      doc.Build.Dockerfile.ImageTarget,
				"testTarget":       doc.Build.Dockerfile.TestTarget,
				"cacheRefTemplate": doc.Build.Dockerfile.CacheRefTemplate,
			},
		}
	}
	if key := strings.TrimSpace(doc.Artifacts.Containerfile); key != "" {
		subset["containerfile"] = map[string]string{"artifactKey": key}
	}
	return subset
}
