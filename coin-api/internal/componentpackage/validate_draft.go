package componentpackage

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type DraftArtifact struct {
	Path   string
	Body   []byte
	SHA256 string
}

type ValidationIssue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidateDraftResult struct {
	Valid  bool              `json:"valid"`
	Issues []ValidationIssue `json:"issues"`
}

type branchingModelDoc struct {
	SchemaVersion int      `yaml:"schemaVersion"`
	Name          string   `yaml:"name"`
	Trunk         struct {
		Branch string `yaml:"branch"`
	} `yaml:"trunk"`
	BranchTypes []string `yaml:"branchTypes"`
	Versioning  struct {
		TagPrefix   string `yaml:"tagPrefix"`
		Qualifiers  struct {
			Snapshot struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"snapshot"`
			RC struct {
				Enabled               bool `yaml:"enabled"`
				ReleaseBranchesOnly   bool `yaml:"releaseBranchesOnly"`
			} `yaml:"rc"`
		} `yaml:"qualifiers"`
	} `yaml:"versioning"`
	Publish struct {
		When   string `yaml:"when"`
		Branch string `yaml:"branch"`
	} `yaml:"publish"`
}

// ValidateDraftPackage runs server-side validation for Component Studio publish flow.
func ValidateDraftPackage(componentType, componentName, version string, bodies []DraftArtifact) ValidateDraftResult {
	var issues []ValidationIssue
	if componentType == "" || componentName == "" || version == "" {
		issues = append(issues, ValidationIssue{Field: "component", Message: "type, name and version are required"})
		return ValidateDraftResult{Valid: false, Issues: issues}
	}
	if len(bodies) == 0 {
		issues = append(issues, ValidationIssue{Field: "artifacts", Message: "draft has no artifact bodies"})
		return ValidateDraftResult{Valid: false, Issues: issues}
	}

	inputs := make([]ArtifactInput, 0, len(bodies))
	for _, b := range bodies {
		if b.Path == "" {
			issues = append(issues, ValidationIssue{Field: "artifacts", Message: "artifact path is required"})
			continue
		}
		if !sha256Pattern.MatchString(b.SHA256) {
			issues = append(issues, ValidationIssue{Field: b.Path, Message: "invalid artifact sha256"})
			continue
		}
		inputs = append(inputs, ArtifactInput{Path: b.Path, SHA256: b.SHA256})
	}
	if _, err := BuildPackageManifestJSON(componentType, componentName, version, inputs); err != nil {
		issues = append(issues, ValidationIssue{Field: "package.manifest", Message: err.Error()})
	}

	switch componentType {
	case "branching-model":
		issues = append(issues, validateBranchingModelArtifact(bodies, componentName)...)
	case "gp-content":
		issues = append(issues, validateGPContentArtifacts(bodies, componentName)...)
	default:
		issues = append(issues, ValidationIssue{
			Field:   "componentType",
			Message: fmt.Sprintf("validation not implemented for type %q", componentType),
		})
	}

	return ValidateDraftResult{Valid: len(issues) == 0, Issues: issues}
}

func validateBranchingModelArtifact(bodies []DraftArtifact, componentName string) []ValidationIssue {
	var issues []ValidationIssue
	var modelRaw []byte
	for _, b := range bodies {
		if b.Path == "model.yaml" {
			modelRaw = b.Body
			break
		}
	}
	if len(modelRaw) == 0 {
		return []ValidationIssue{{Field: "model.yaml", Message: "required primary artifact is missing"}}
	}

	var doc branchingModelDoc
	if err := yaml.Unmarshal(modelRaw, &doc); err != nil {
		return []ValidationIssue{{Field: "model.yaml", Message: fmt.Sprintf("invalid yaml: %v", err)}}
	}
	if doc.SchemaVersion != 1 {
		issues = append(issues, ValidationIssue{Field: "model.yaml.schemaVersion", Message: "schemaVersion must be 1"})
	}
	if strings.TrimSpace(doc.Name) == "" {
		issues = append(issues, ValidationIssue{Field: "model.yaml.name", Message: "name is required"})
	} else if doc.Name != componentName {
		issues = append(issues, ValidationIssue{
			Field:   "model.yaml.name",
			Message: fmt.Sprintf("name %q must match component name %q", doc.Name, componentName),
		})
	}
	if strings.TrimSpace(doc.Trunk.Branch) == "" {
		issues = append(issues, ValidationIssue{Field: "model.yaml.trunk.branch", Message: "trunk branch is required"})
	}
	if len(doc.BranchTypes) == 0 {
		issues = append(issues, ValidationIssue{Field: "model.yaml.branchTypes", Message: "at least one branch type is required"})
	} else {
		hasRelease := false
		for _, t := range doc.BranchTypes {
			if t == "release" {
				hasRelease = true
			}
		}
		if !hasRelease {
			issues = append(issues, ValidationIssue{Field: "model.yaml.branchTypes", Message: "branchTypes must include release"})
		}
	}
	switch doc.Publish.When {
	case "tag", "branch", "always", "never":
		if doc.Publish.When == "branch" && strings.TrimSpace(doc.Publish.Branch) == "" {
			issues = append(issues, ValidationIssue{Field: "model.yaml.publish.branch", Message: "branch is required when publish.when is branch"})
		}
	default:
		issues = append(issues, ValidationIssue{
			Field:   "model.yaml.publish.when",
			Message: "publish.when must be tag, branch, always or never",
		})
	}
	return issues
}

type gpContentDoc struct {
	Name  string `yaml:"name"`
	Kind  string `yaml:"kind"`
	Build struct {
		Engine string `yaml:"engine"`
	} `yaml:"build"`
	Pipeline struct {
		Stages []struct {
			ID   string `yaml:"id"`
			Name string `yaml:"name"`
		} `yaml:"stages"`
	} `yaml:"pipeline"`
	ValidateSchema struct {
		ArtifactKey string `yaml:"artifactKey"`
	} `yaml:"validateSchema"`
	Containerfile struct {
		ArtifactKey string `yaml:"artifactKey"`
	} `yaml:"containerfile"`
}

func validateGPContentArtifacts(bodies []DraftArtifact, componentName string) []ValidationIssue {
	var issues []ValidationIssue
	var contentRaw, containerRaw []byte
	for _, b := range bodies {
		switch b.Path {
		case "content.yaml":
			contentRaw = b.Body
		case "dockerfiles/Containerfile":
			containerRaw = b.Body
		}
	}
	if len(contentRaw) == 0 {
		issues = append(issues, ValidationIssue{Field: "content.yaml", Message: "required primary artifact is missing"})
		return issues
	}
	if len(containerRaw) == 0 {
		issues = append(issues, ValidationIssue{Field: "dockerfiles/Containerfile", Message: "required containerfile artifact is missing"})
	}

	var doc gpContentDoc
	if err := yaml.Unmarshal(contentRaw, &doc); err != nil {
		return append(issues, ValidationIssue{Field: "content.yaml", Message: fmt.Sprintf("invalid yaml: %v", err)})
	}
	if strings.TrimSpace(doc.Name) == "" {
		issues = append(issues, ValidationIssue{Field: "content.yaml.name", Message: "name is required"})
	} else if doc.Name != componentName {
		issues = append(issues, ValidationIssue{
			Field:   "content.yaml.name",
			Message: fmt.Sprintf("name %q must match component name %q", doc.Name, componentName),
		})
	}
	if doc.Kind != "gp-content" {
		issues = append(issues, ValidationIssue{Field: "content.yaml.kind", Message: "kind must be gp-content"})
	}
	switch doc.Build.Engine {
	case "buildkit", "buildpack", "dockerfile":
	default:
		issues = append(issues, ValidationIssue{Field: "content.yaml.build.engine", Message: "build.engine must be buildkit, buildpack or dockerfile"})
	}
	if len(doc.Pipeline.Stages) == 0 {
		issues = append(issues, ValidationIssue{Field: "content.yaml.pipeline.stages", Message: "at least one pipeline stage is required"})
	} else {
		for i, st := range doc.Pipeline.Stages {
			if strings.TrimSpace(st.ID) == "" || strings.TrimSpace(st.Name) == "" {
				issues = append(issues, ValidationIssue{
					Field:   fmt.Sprintf("content.yaml.pipeline.stages[%d]", i),
					Message: "stage id and name are required",
				})
			}
		}
	}
	if strings.TrimSpace(doc.ValidateSchema.ArtifactKey) == "" {
		issues = append(issues, ValidationIssue{Field: "content.yaml.validateSchema.artifactKey", Message: "validateSchema.artifactKey is required"})
	}
	if strings.TrimSpace(doc.Containerfile.ArtifactKey) == "" {
		issues = append(issues, ValidationIssue{Field: "content.yaml.containerfile.artifactKey", Message: "containerfile.artifactKey is required"})
	}
	return issues
}
