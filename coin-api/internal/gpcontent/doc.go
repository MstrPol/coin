package gpcontent

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const SchemaVersion = 2

type Doc struct {
	SchemaVersion  int           `yaml:"schemaVersion"`
	Name           string        `yaml:"name"`
	Kind           string        `yaml:"kind"`
	Version        string        `yaml:"-"`
	ValidateSchema string        `yaml:"validateSchema,omitempty"`
	Capabilities   Capabilities  `yaml:"capabilities"`
	Parameters     []Parameter   `yaml:"parameters,omitempty"`
	Build          Build         `yaml:"build"`
	Deliverables   []Deliverable `yaml:"deliverables,omitempty"`
	Pipeline       Pipeline      `yaml:"pipeline"`
	Artifacts      Artifacts     `yaml:"artifacts"`
}

type Parameter struct {
	Name          string   `yaml:"name"`
	Type          string   `yaml:"type"`
	Default       any      `yaml:"default,omitempty"`
	Required      bool     `yaml:"required,omitempty"`
	AllowedValues []string `yaml:"allowedValues,omitempty"`
}

type Capabilities struct {
	Deliverables []string `yaml:"deliverables"`
}

type Build struct {
	Engine     string           `yaml:"engine"`
	Buildkit   *BuildkitBlock   `yaml:"buildkit,omitempty"`
	Dockerfile *DockerfileBlock `yaml:"dockerfile,omitempty"`
	Targets    []BuildTarget    `yaml:"targets,omitempty"`
}

type BuildkitBlock struct {
	Targets map[string]string `yaml:"targets"`
}

type DockerfileBlock struct {
	Path        string `yaml:"path"`
	ImageTarget string `yaml:"imageTarget,omitempty"`
	TestTarget  string `yaml:"testTarget,omitempty"`
}

type BuildTarget struct {
	ID            string `yaml:"id"`
	Engine        string `yaml:"engine"`
	Containerfile string `yaml:"containerfile,omitempty"`
	Target        string `yaml:"target,omitempty"`
	Dockerfile    string `yaml:"dockerfile,omitempty"`
	Output        string `yaml:"output,omitempty"`
}

type Deliverable struct {
	ID       string               `yaml:"id"`
	Type     string               `yaml:"type"`
	TargetID string               `yaml:"targetId"`
	Image    *ImageDeliverable    `yaml:"image,omitempty"`
	Artifact *ArtifactDeliverable `yaml:"artifact,omitempty"`
}

type ImageDeliverable struct {
	RepositorySuffix string `yaml:"repositorySuffix,omitempty"`
}

type ArtifactDeliverable struct {
	Format string   `yaml:"format,omitempty"`
	Paths  []string `yaml:"paths,omitempty"`
}

type Pipeline struct {
	Stages []Stage `yaml:"stages"`
}

type Stage struct {
	ID    string      `yaml:"id"`
	Name  string      `yaml:"name"`
	Steps []StageStep `yaml:"steps,omitempty"`
}

type StageStep struct {
	Action        string         `yaml:"action"`
	TargetID      string         `yaml:"targetId,omitempty"`
	DeliverableID string         `yaml:"deliverableId,omitempty"`
	Run           *InlineRun     `yaml:"run,omitempty"`
	Build         *InlineBuild   `yaml:"build,omitempty"`
	Publish       *InlinePublish `yaml:"publish,omitempty"`
}

type Artifacts struct {
	ValidateSchema string              `yaml:"validateSchema"`
	Containerfile  string              `yaml:"containerfile,omitempty"`
	Containerfiles []ContainerfileSpec `yaml:"containerfiles,omitempty"`
}

type ContainerfileSpec struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

type Issue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type PreviewResult struct {
	Valid        bool           `json:"valid"`
	Issues       []Issue        `json:"issues"`
	Warnings     []string       `json:"warnings,omitempty"`
	Parameters   []Parameter    `json:"parameters,omitempty"`
	Build        map[string]any `json:"build,omitempty"`
	Deliverables []Deliverable  `json:"deliverables,omitempty"`
	Pipeline     map[string]any `json:"pipeline,omitempty"`
	Artifacts    map[string]any `json:"artifacts,omitempty"`
	Capabilities map[string]any `json:"capabilities,omitempty"`
}

type PreviewOptions struct {
	ComponentName            string
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
	if isInlineDoc(doc) {
		if strings.TrimSpace(doc.Name) == "" {
			issues = append(issues, Issue{Field: "content.yaml.name", Message: "name is required"})
		} else if name != "" && doc.Name != name {
			issues = append(issues, Issue{
				Field:   "content.yaml.name",
				Message: fmt.Sprintf("name %q must match component name %q", doc.Name, name),
			})
		}
		if doc.Kind != "gp-content" && doc.Kind != "golden-path" {
			issues = append(issues, Issue{Field: "content.yaml.kind", Message: "kind must be golden-path"})
		}
		issues = append(issues, validateInlineDoc(doc)...)
		return issues, warnings
	}
	if doc.SchemaVersion == SchemaVersion && isVNextDoc(doc) {
		issues = append(issues, Issue{
			Field:   "content.yaml.schemaVersion",
			Message: "v2 catalog model is superseded; use schemaVersion 3 pipeline-inline",
		})
		return issues, warnings
	}
	if doc.SchemaVersion != SchemaVersion {
		issues = append(issues, Issue{Field: "content.yaml.schemaVersion", Message: "schemaVersion must be 2 (legacy flat) or 3 (pipeline-inline); v1 is not supported"})
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
	if isVNextDoc(doc) {
		issues = append(issues, validateVNextDoc(doc)...)
		if strings.TrimSpace(doc.Artifacts.ValidateSchema) == "" {
			issues = append(issues, Issue{Field: "content.yaml.artifacts.validateSchema", Message: "validateSchema artifact is required"})
		}
		return issues, warnings
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
	allowedDeliverables := map[string]struct{}{
		"image":           {},
		"liquibase-image": {},
		"artifact":        {},
	}
	seenDeliverables := map[string]struct{}{}
	for _, d := range doc.Capabilities.Deliverables {
		if _, ok := allowedDeliverables[d]; !ok {
			issues = append(issues, Issue{Field: "content.yaml.capabilities.deliverables", Message: fmt.Sprintf("deliverable %q is not supported in P0", d)})
			continue
		}
		if _, dup := seenDeliverables[d]; dup {
			issues = append(issues, Issue{Field: "content.yaml.capabilities.deliverables", Message: fmt.Sprintf("duplicate deliverable type %q", d)})
		}
		seenDeliverables[d] = struct{}{}
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

func isVNextDoc(doc Doc) bool {
	if isInlineDoc(doc) {
		return false
	}
	if len(doc.Parameters) > 0 || len(doc.Build.Targets) > 0 || len(doc.Deliverables) > 0 || len(doc.Artifacts.Containerfiles) > 0 {
		return true
	}
	for _, stage := range doc.Pipeline.Stages {
		if len(stage.Steps) > 0 {
			return true
		}
	}
	return false
}

func validateVNextDoc(doc Doc) []Issue {
	var issues []Issue
	parameterNames := map[string]struct{}{}
	for i, p := range doc.Parameters {
		prefix := fmt.Sprintf("content.yaml.parameters[%d]", i)
		name := strings.TrimSpace(p.Name)
		if name == "" {
			issues = append(issues, Issue{Field: prefix + ".name", Message: "parameter name is required"})
		} else {
			if _, dup := parameterNames[name]; dup {
				issues = append(issues, Issue{Field: prefix + ".name", Message: fmt.Sprintf("duplicate parameter %q", name)})
			}
			parameterNames[name] = struct{}{}
			if looksSecretParameter(name) {
				issues = append(issues, Issue{Field: prefix + ".name", Message: "parameters must not carry secrets or Jenkins credential IDs"})
			}
		}
		switch strings.TrimSpace(p.Type) {
		case "string", "boolean", "number":
		case "enum":
			if len(p.AllowedValues) == 0 {
				issues = append(issues, Issue{Field: prefix + ".allowedValues", Message: "enum parameter requires allowedValues"})
			}
		default:
			issues = append(issues, Issue{Field: prefix + ".type", Message: "parameter type must be string, boolean, number or enum"})
		}
		if p.Required && p.Default == nil {
			issues = append(issues, Issue{Field: prefix + ".default", Message: "required parameter must define default value"})
		}
	}

	containerfiles := map[string]struct{}{}
	for i, cf := range doc.Artifacts.Containerfiles {
		prefix := fmt.Sprintf("content.yaml.artifacts.containerfiles[%d]", i)
		id := strings.TrimSpace(cf.ID)
		if id == "" {
			issues = append(issues, Issue{Field: prefix + ".id", Message: "containerfile id is required"})
		} else {
			if _, dup := containerfiles[id]; dup {
				issues = append(issues, Issue{Field: prefix + ".id", Message: fmt.Sprintf("duplicate containerfile %q", id)})
			}
			containerfiles[id] = struct{}{}
		}
		if strings.TrimSpace(cf.Path) == "" {
			issues = append(issues, Issue{Field: prefix + ".path", Message: "containerfile path is required"})
		}
	}

	targets := map[string]BuildTarget{}
	for i, target := range doc.Build.Targets {
		prefix := fmt.Sprintf("content.yaml.build.targets[%d]", i)
		id := strings.TrimSpace(target.ID)
		if id == "" {
			issues = append(issues, Issue{Field: prefix + ".id", Message: "target id is required"})
		} else {
			if _, dup := targets[id]; dup {
				issues = append(issues, Issue{Field: prefix + ".id", Message: fmt.Sprintf("duplicate target %q", id)})
			}
			targets[id] = target
		}
		switch strings.TrimSpace(target.Engine) {
		case "buildkit":
			cf := strings.TrimSpace(target.Containerfile)
			if cf == "" {
				issues = append(issues, Issue{Field: prefix + ".containerfile", Message: "buildkit target requires containerfile"})
			} else if _, ok := containerfiles[cf]; !ok {
				issues = append(issues, Issue{Field: prefix + ".containerfile", Message: fmt.Sprintf("containerfile %q is not declared", cf)})
			}
			if strings.TrimSpace(target.Target) == "" {
				issues = append(issues, Issue{Field: prefix + ".target", Message: "buildkit target requires target"})
			}
		case "dockerfile":
			if strings.TrimSpace(target.Dockerfile) == "" {
				issues = append(issues, Issue{Field: prefix + ".dockerfile", Message: "dockerfile target requires dockerfile path"})
			}
		default:
			issues = append(issues, Issue{Field: prefix + ".engine", Message: "target engine must be buildkit or dockerfile"})
		}
	}
	if len(doc.Build.Targets) == 0 {
		issues = append(issues, Issue{Field: "content.yaml.build.targets", Message: "at least one build target is required"})
	}

	deliverables := map[string]Deliverable{}
	for i, deliverable := range doc.Deliverables {
		prefix := fmt.Sprintf("content.yaml.deliverables[%d]", i)
		id := strings.TrimSpace(deliverable.ID)
		if id == "" {
			issues = append(issues, Issue{Field: prefix + ".id", Message: "deliverable id is required"})
		} else {
			if _, dup := deliverables[id]; dup {
				issues = append(issues, Issue{Field: prefix + ".id", Message: fmt.Sprintf("duplicate deliverable %q", id)})
			}
			deliverables[id] = deliverable
		}
		switch strings.TrimSpace(deliverable.Type) {
		case "image", "liquibase-image", "artifact":
		default:
			issues = append(issues, Issue{Field: prefix + ".type", Message: "deliverable type must be image, liquibase-image or artifact"})
		}
		targetID := strings.TrimSpace(deliverable.TargetID)
		if targetID == "" {
			issues = append(issues, Issue{Field: prefix + ".targetId", Message: "targetId is required"})
		} else if _, ok := targets[targetID]; !ok {
			issues = append(issues, Issue{Field: prefix + ".targetId", Message: fmt.Sprintf("target %q is not declared", targetID)})
		}
	}
	if len(doc.Deliverables) == 0 {
		issues = append(issues, Issue{Field: "content.yaml.deliverables", Message: "at least one deliverable is required"})
	}

	if len(doc.Pipeline.Stages) == 0 {
		issues = append(issues, Issue{Field: "content.yaml.pipeline.stages", Message: "at least one pipeline stage is required"})
	}
	for i, stage := range doc.Pipeline.Stages {
		prefix := fmt.Sprintf("content.yaml.pipeline.stages[%d]", i)
		if strings.TrimSpace(stage.ID) == "" || strings.TrimSpace(stage.Name) == "" {
			issues = append(issues, Issue{Field: prefix, Message: "stage id and name are required"})
		}
		if len(stage.Steps) == 0 {
			issues = append(issues, Issue{Field: prefix + ".steps", Message: "stage requires at least one step"})
		}
		for j, step := range stage.Steps {
			stepPrefix := fmt.Sprintf("%s.steps[%d]", prefix, j)
			switch strings.TrimSpace(step.Action) {
			case "run-target":
				if _, ok := targets[strings.TrimSpace(step.TargetID)]; !ok {
					issues = append(issues, Issue{Field: stepPrefix + ".targetId", Message: "run-target step requires declared targetId"})
				}
			case "build-deliverable", "publish-deliverable":
				if _, ok := deliverables[strings.TrimSpace(step.DeliverableID)]; !ok {
					issues = append(issues, Issue{Field: stepPrefix + ".deliverableId", Message: fmt.Sprintf("%s step requires declared deliverableId", step.Action)})
				}
			default:
				issues = append(issues, Issue{Field: stepPrefix + ".action", Message: "step action must be run-target, build-deliverable or publish-deliverable"})
			}
		}
	}

	return issues
}

func looksSecretParameter(name string) bool {
	upper := strings.ToUpper(name)
	secretMarkers := []string{"SECRET", "PASSWORD", "PASSWD", "TOKEN", "CREDENTIAL", "CREDENTIALS"}
	for _, marker := range secretMarkers {
		if strings.Contains(upper, marker) {
			return true
		}
	}
	return false
}

func Preview(doc Doc, opts PreviewOptions) PreviewResult {
	if isInlineDoc(doc) {
		if doc.Version == "" {
			doc.Version = "draft"
		}
		return previewInlineDoc(doc, opts)
	}
	issues, warnings := ValidateDoc(doc, opts)
	out := PreviewResult{
		Valid:    len(issues) == 0,
		Issues:   issues,
		Warnings: warnings,
	}
	if isVNextDoc(doc) {
		out.Parameters = doc.Parameters
		out.Deliverables = doc.Deliverables
		out.Build = map[string]any{"targets": doc.Build.Targets}
		out.Pipeline = map[string]any{"stages": doc.Pipeline.Stages}
		if len(doc.Artifacts.Containerfiles) > 0 {
			containerfiles := make([]map[string]string, 0, len(doc.Artifacts.Containerfiles))
			for _, cf := range doc.Artifacts.Containerfiles {
				containerfiles = append(containerfiles, map[string]string{
					"id":          cf.ID,
					"artifactKey": cf.Path,
				})
			}
			out.Artifacts = map[string]any{"containerfiles": containerfiles}
		}
		return out
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
	switch engine {
	case "buildkit":
		if doc.Build.Buildkit != nil {
			bk := map[string]any{
				"dockerfile": ".coin/Containerfile",
				"targets":    doc.Build.Buildkit.Targets,
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
			buildDoc["dockerfile"] = df
		}
	}
	out.Build = buildDoc
	return out
}

func ManifestSubset(doc Doc) map[string]any {
	if isInlineDoc(doc) {
		return manifestSubsetInline(doc)
	}
	if isVNextDoc(doc) {
		subset := map[string]any{
			"parameters": doc.Parameters,
			"build": map[string]any{
				"targets": doc.Build.Targets,
			},
			"deliverables": doc.Deliverables,
			"pipeline": map[string]any{
				"stages": doc.Pipeline.Stages,
			},
			"validateSchema": map[string]string{
				"artifactKey": doc.Artifacts.ValidateSchema,
			},
		}
		if len(doc.Artifacts.Containerfiles) > 0 {
			containerfiles := make([]map[string]string, 0, len(doc.Artifacts.Containerfiles))
			for _, cf := range doc.Artifacts.Containerfiles {
				containerfiles = append(containerfiles, map[string]string{
					"id":          cf.ID,
					"artifactKey": cf.Path,
				})
			}
			subset["artifacts"] = map[string]any{
				"containerfiles": containerfiles,
			}
		}
		return subset
	}
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
				"targets": doc.Build.Buildkit.Targets,
			},
		}
	}
	if doc.Build.Dockerfile != nil {
		subset["build"] = map[string]any{
			"engine": doc.Build.Engine,
			"dockerfile": map[string]any{
				"path":        doc.Build.Dockerfile.Path,
				"imageTarget": doc.Build.Dockerfile.ImageTarget,
				"testTarget":  doc.Build.Dockerfile.TestTarget,
			},
		}
	}
	if key := strings.TrimSpace(doc.Artifacts.Containerfile); key != "" {
		subset["containerfile"] = map[string]string{"artifactKey": key}
	}
	return subset
}
