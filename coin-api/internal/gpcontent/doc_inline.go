package gpcontent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

const SchemaVersionInline = 3

var shortIDPattern = regexp.MustCompile(`^[a-z0-9]{5,6}$`)

func validateShortID(field, id string) []Issue {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	if !shortIDPattern.MatchString(id) {
		return []Issue{{
			Field:   field,
			Message: "id must be 5-6 lowercase alphanumeric characters (short hash)",
		}}
	}
	return nil
}

type InlineContainerfile struct {
	Body       string         `yaml:"body,omitempty" json:"body,omitempty"`
	ContentRef map[string]any `yaml:"contentRef,omitempty" json:"contentRef,omitempty"`
	Digest     string         `yaml:"digest,omitempty" json:"digest,omitempty"`
}

type InlineDockerfile struct {
	Path   string `yaml:"path" json:"path"`
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
}

type InlineRun struct {
	Engine        string               `yaml:"engine" json:"engine"`
	Target        string               `yaml:"target,omitempty" json:"target,omitempty"`
	Output        string               `yaml:"output,omitempty" json:"output,omitempty"`
	Containerfile *InlineContainerfile `yaml:"containerfile,omitempty" json:"containerfile,omitempty"`
	Dockerfile    *InlineDockerfile    `yaml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
}

type InlineBuild struct {
	ID            string               `yaml:"id" json:"id"`
	Type          string               `yaml:"type" json:"type"`
	Engine        string               `yaml:"engine" json:"engine"`
	Target        string               `yaml:"target,omitempty" json:"target,omitempty"`
	Containerfile *InlineContainerfile `yaml:"containerfile,omitempty" json:"containerfile,omitempty"`
	Dockerfile    *InlineDockerfile    `yaml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	Image         *ImageDeliverable    `yaml:"image,omitempty" json:"image,omitempty"`
	Artifact      *ArtifactDeliverable `yaml:"artifact,omitempty" json:"artifact,omitempty"`
}

type InlinePublish struct {
	BuildStepID string `yaml:"buildStepId" json:"buildStepId"`
}

func isInlineDoc(doc Doc) bool {
	return doc.SchemaVersion == SchemaVersionInline
}

func hasVNextCatalog(doc Doc) bool {
	return len(doc.Build.Targets) > 0 || len(doc.Deliverables) > 0 || len(doc.Artifacts.Containerfiles) > 0
}

func validateInlineDoc(doc Doc) []Issue {
	var issues []Issue
	if hasVNextCatalog(doc) {
		issues = append(issues, Issue{
			Field:   "content.yaml",
			Message: "schemaVersion 3 must not include build.targets, deliverables or artifacts.containerfiles",
		})
	}
	schemaPath := strings.TrimSpace(doc.ValidateSchema)
	if schemaPath == "" {
		schemaPath = strings.TrimSpace(doc.Artifacts.ValidateSchema)
	}
	if schemaPath == "" {
		issues = append(issues, Issue{Field: "content.yaml.validateSchema", Message: "validateSchema is required"})
	}

	issues = append(issues, validateParameters(doc)...)

	buildIDs := map[string]struct{}{}
	for i, stage := range doc.Pipeline.Stages {
		prefix := fmt.Sprintf("content.yaml.pipeline.stages[%d]", i)
		if strings.TrimSpace(stage.ID) == "" || strings.TrimSpace(stage.Name) == "" {
			issues = append(issues, Issue{Field: prefix, Message: "stage id and name are required"})
		}
		issues = append(issues, validateShortID(prefix+".id", stage.ID)...)
		if len(stage.Steps) == 0 {
			issues = append(issues, Issue{Field: prefix + ".steps", Message: "stage requires at least one step"})
		}
		for j, step := range stage.Steps {
			stepPrefix := fmt.Sprintf("%s.steps[%d]", prefix, j)
			switch strings.TrimSpace(step.Action) {
			case "run":
				issues = append(issues, validateInlineRun(stepPrefix+".run", step.Run)...)
			case "build":
				if step.Build == nil {
					issues = append(issues, Issue{Field: stepPrefix + ".build", Message: "build block is required"})
					continue
				}
				id := strings.TrimSpace(step.Build.ID)
				if id == "" {
					issues = append(issues, Issue{Field: stepPrefix + ".build.id", Message: "build.id is required"})
				} else {
					issues = append(issues, validateShortID(stepPrefix+".build.id", id)...)
					if _, dup := buildIDs[id]; dup {
						issues = append(issues, Issue{Field: stepPrefix + ".build.id", Message: fmt.Sprintf("duplicate build.id %q", id)})
					} else {
						buildIDs[id] = struct{}{}
					}
				}
				switch strings.TrimSpace(step.Build.Type) {
				case "image", "liquibase-image", "artifact":
				default:
					issues = append(issues, Issue{Field: stepPrefix + ".build.type", Message: "build.type must be image, liquibase-image or artifact"})
				}
				issues = append(issues, validateInlineEngine(stepPrefix+".build", step.Build.Engine, step.Build.Containerfile, step.Build.Dockerfile)...)
			case "publish":
				if step.Publish == nil || strings.TrimSpace(step.Publish.BuildStepID) == "" {
					issues = append(issues, Issue{Field: stepPrefix + ".publish.buildStepId", Message: "publish.buildStepId is required"})
				} else if _, ok := buildIDs[strings.TrimSpace(step.Publish.BuildStepID)]; !ok {
					issues = append(issues, Issue{Field: stepPrefix + ".publish.buildStepId", Message: fmt.Sprintf("build step %q not found", step.Publish.BuildStepID)})
				}
			default:
				issues = append(issues, Issue{Field: stepPrefix + ".action", Message: "step action must be run, build or publish"})
			}
		}
	}
	if len(doc.Pipeline.Stages) == 0 {
		issues = append(issues, Issue{Field: "content.yaml.pipeline.stages", Message: "at least one pipeline stage is required"})
	}
	return issues
}

func validateParameters(doc Doc) []Issue {
	var issues []Issue
	names := map[string]struct{}{}
	for i, p := range doc.Parameters {
		prefix := fmt.Sprintf("content.yaml.parameters[%d]", i)
		name := strings.TrimSpace(p.Name)
		if name == "" {
			issues = append(issues, Issue{Field: prefix + ".name", Message: "parameter name is required"})
		} else {
			if _, dup := names[name]; dup {
				issues = append(issues, Issue{Field: prefix + ".name", Message: fmt.Sprintf("duplicate parameter %q", name)})
			}
			names[name] = struct{}{}
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
	return issues
}

func validateInlineRun(prefix string, run *InlineRun) []Issue {
	if run == nil {
		return []Issue{{Field: prefix, Message: "run block is required"}}
	}
	return validateInlineEngine(prefix, run.Engine, run.Containerfile, run.Dockerfile)
}

func validateInlineEngine(prefix, engine string, cf *InlineContainerfile, df *InlineDockerfile) []Issue {
	var issues []Issue
	switch strings.TrimSpace(engine) {
	case "buildkit":
		if cf == nil || strings.TrimSpace(cf.Body) == "" {
			issues = append(issues, Issue{Field: prefix + ".containerfile.body", Message: "buildkit step requires containerfile.body"})
		}
	case "dockerfile":
		if df == nil || strings.TrimSpace(df.Path) == "" {
			issues = append(issues, Issue{Field: prefix + ".dockerfile.path", Message: "dockerfile engine requires dockerfile.path"})
		}
	default:
		issues = append(issues, Issue{Field: prefix + ".engine", Message: "engine must be buildkit or dockerfile"})
	}
	return issues
}

func digestContainerfileBody(body string) string {
	sum := sha256.Sum256([]byte(body))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func materializeInlineStages(stages []Stage, componentName, componentVersion string) []Stage {
	out := make([]Stage, len(stages))
	for i, stage := range stages {
		out[i] = Stage{
			ID:    stage.ID,
			Name:  stage.Name,
			Steps: materializeInlineSteps(stage.Steps, componentName, componentVersion),
		}
	}
	return out
}

func materializeInlineSteps(steps []StageStep, componentName, componentVersion string) []StageStep {
	out := make([]StageStep, len(steps))
	for i, step := range steps {
		out[i] = StageStep{
			Action:        step.Action,
			TargetID:      step.TargetID,
			DeliverableID: step.DeliverableID,
			Run:           materializeInlineRun(step.Run, componentName, componentVersion),
			Build:         materializeInlineBuild(step.Build, componentName, componentVersion),
			Publish:       step.Publish,
		}
	}
	return out
}

func materializeInlineRun(run *InlineRun, componentName, componentVersion string) *InlineRun {
	if run == nil {
		return nil
	}
	cp := *run
	if cp.Containerfile != nil && strings.TrimSpace(cp.Containerfile.Body) != "" {
		digest := digestContainerfileBody(cp.Containerfile.Body)
		cp.Containerfile = &InlineContainerfile{
			ContentRef: map[string]any{
				"url":    ContentArtifactURL(componentName, componentVersion, "containerfile:"+digest),
				"sha256": digest,
			},
			Digest: digest,
		}
	}
	return &cp
}

func materializeInlineBuild(build *InlineBuild, componentName, componentVersion string) *InlineBuild {
	if build == nil {
		return nil
	}
	cp := *build
	if cp.Containerfile != nil && strings.TrimSpace(cp.Containerfile.Body) != "" {
		digest := digestContainerfileBody(cp.Containerfile.Body)
		cp.Containerfile = &InlineContainerfile{
			ContentRef: map[string]any{
				"url":    ContentArtifactURL(componentName, componentVersion, "containerfile:"+digest),
				"sha256": digest,
			},
			Digest: digest,
		}
	}
	return &cp
}

// ContentArtifactURL builds artifact URL for embedded GP release pipeline materialization.
func ContentArtifactURL(gpName, gpVersion, artifactKey string) string {
	return fmt.Sprintf("coin://golden-path/%s@%s/%s", gpName, gpVersion, artifactKey)
}

func previewInlineDoc(doc Doc, opts PreviewOptions) PreviewResult {
	issues, warnings := ValidateDoc(doc, opts)
	out := PreviewResult{
		Valid:      len(issues) == 0,
		Issues:     issues,
		Warnings:   warnings,
		Parameters: doc.Parameters,
		Pipeline: map[string]any{
			"stages": materializeInlineStages(doc.Pipeline.Stages, doc.Name, doc.Version),
		},
	}
	schemaPath := strings.TrimSpace(doc.ValidateSchema)
	if schemaPath == "" {
		schemaPath = strings.TrimSpace(doc.Artifacts.ValidateSchema)
	}
	if schemaPath != "" {
		out.Artifacts = map[string]any{
			"validateSchema": map[string]string{"artifactKey": schemaPath},
		}
	}
	return out
}

func manifestSubsetInline(doc Doc) map[string]any {
	schemaPath := strings.TrimSpace(doc.ValidateSchema)
	if schemaPath == "" {
		schemaPath = strings.TrimSpace(doc.Artifacts.ValidateSchema)
	}
	name := doc.Name
	version := doc.Version
	return map[string]any{
		"schemaVersion": SchemaVersionInline,
		"parameters":    doc.Parameters,
		"pipeline": map[string]any{
			"stages": materializeInlineStages(doc.Pipeline.Stages, name, version),
		},
		"validateSchema": map[string]string{
			"artifactKey": schemaPath,
		},
	}
}
