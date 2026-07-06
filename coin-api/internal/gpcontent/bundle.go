package gpcontent

import (
	"encoding/json"

	"coin.local/coin-api/internal/manifest"
)

// ContentBundleFromInlineDoc builds a resolve ContentBundle from pipeline-inline Doc.
func ContentBundleFromInlineDoc(doc Doc, gpName, gpVersion string) manifest.ContentBundle {
	materialized := materializeInlineStages(doc.Pipeline.Stages, gpName, gpVersion)
	stages := make([]manifest.TypedStage, 0, len(materialized))
	for _, st := range materialized {
		stages = append(stages, manifestStageFromInline(st))
	}
	params := make([]manifest.Parameter, 0, len(doc.Parameters))
	for _, p := range doc.Parameters {
		params = append(params, manifest.Parameter{
			Name:          p.Name,
			Type:          p.Type,
			Default:       p.Default,
			Required:      p.Required,
			AllowedValues: p.AllowedValues,
		})
	}
	schemaKey := doc.ValidateSchema
	if schemaKey == "" {
		schemaKey = doc.Artifacts.ValidateSchema
	}
	return manifest.ContentBundle{
		Name:              gpName,
		Version:           gpVersion,
		Parameters:        params,
		SchemaArtifactKey: schemaKey,
		Stages:            stages,
	}
}

func manifestStageFromInline(st Stage) manifest.TypedStage {
	out := manifest.TypedStage{
		ID:   st.ID,
		Name: st.Name,
	}
	if len(st.Steps) > 0 {
		out.Steps = make([]manifest.StageStep, 0, len(st.Steps))
		for _, step := range st.Steps {
			out.Steps = append(out.Steps, manifestStepFromInline(step))
		}
	}
	return out
}

func manifestStepFromInline(step StageStep) manifest.StageStep {
	out := manifest.StageStep{Action: step.Action}
	if step.Run != nil {
		out.Run = &manifest.InlineRunStep{
			Engine: step.Run.Engine,
			Target: step.Run.Target,
			Output: step.Run.Output,
		}
		if step.Run.Containerfile != nil {
			out.Run.Containerfile = inlineContainerfileToManifest(step.Run.Containerfile)
		}
		if step.Run.Dockerfile != nil {
			out.Run.Dockerfile = &manifest.InlineDockerfileStep{
				Path:   step.Run.Dockerfile.Path,
				Target: step.Run.Dockerfile.Target,
			}
		}
	}
	if step.Build != nil {
		out.Build = &manifest.InlineBuildStep{
			ID:     step.Build.ID,
			Type:   step.Build.Type,
			Engine: step.Build.Engine,
			Target: step.Build.Target,
		}
		if step.Build.Containerfile != nil {
			out.Build.Containerfile = inlineContainerfileToManifest(step.Build.Containerfile)
		}
		if step.Build.Dockerfile != nil {
			out.Build.Dockerfile = &manifest.InlineDockerfileStep{
				Path:   step.Build.Dockerfile.Path,
				Target: step.Build.Dockerfile.Target,
			}
		}
		if step.Build.Image != nil {
			out.Build.Image = &manifest.ImageDeliverable{RepositorySuffix: step.Build.Image.RepositorySuffix}
		}
		if step.Build.Artifact != nil {
			out.Build.Artifact = &manifest.ArtifactDeliverable{
				Format: step.Build.Artifact.Format,
				Paths:  step.Build.Artifact.Paths,
			}
		}
	}
	if step.Publish != nil {
		out.Publish = &manifest.InlinePublishStep{BuildStepID: step.Publish.BuildStepID}
	}
	return out
}

func inlineContainerfileToManifest(cf *InlineContainerfile) *manifest.InlineContainerfileStep {
	if cf == nil {
		return nil
	}
	out := &manifest.InlineContainerfileStep{Body: cf.Body, Digest: cf.Digest}
	if cf.ContentRef != nil {
		out.ContentRef = make(map[string]string, len(cf.ContentRef))
		for k, v := range cf.ContentRef {
			if s, ok := v.(string); ok {
				out.ContentRef[k] = s
			}
		}
	}
	return out
}

// PipelineBodyJSON returns canonical JSON bytes for storage in gp_release_pipeline_bodies.
func PipelineBodyJSON(doc Doc) ([]byte, error) {
	payload := map[string]any{
		"schemaVersion":  doc.SchemaVersion,
		"parameters":     doc.Parameters,
		"validateSchema": firstNonEmpty(doc.ValidateSchema, doc.Artifacts.ValidateSchema),
		"pipeline":       doc.Pipeline,
	}
	return json.Marshal(payload)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
