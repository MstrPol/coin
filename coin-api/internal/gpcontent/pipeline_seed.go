package gpcontent

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed seed/pipelines/*.yaml
var pipelineSeedFS embed.FS

// DefaultPipelineDoc returns embedded pipeline-inline Doc for a GP profile name.
func DefaultPipelineDoc(gpName string) (Doc, error) {
	name := strings.TrimSpace(gpName)
	if name == "" {
		return Doc{}, fmt.Errorf("gp profile name is required")
	}
	raw, err := pipelineSeedFS.ReadFile("seed/pipelines/" + name + ".yaml")
	if err != nil {
		return Doc{}, fmt.Errorf("no embedded pipeline for profile %q: %w", name, err)
	}
	doc, err := ParseDoc(raw)
	if err != nil {
		return Doc{}, err
	}
	doc.Name = name
	doc.Kind = "golden-path"
	return doc, nil
}

// SeedPipelineBody inserts default embedded pipeline for a GP release when none exists.
func SeedPipelineBody(gpName string, gpVersion string) ([]byte, error) {
	doc, err := DefaultPipelineDoc(gpName)
	if err != nil {
		return nil, err
	}
	doc.Version = gpVersion
	return PipelineBodyJSON(doc)
}
