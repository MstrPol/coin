package store

import (
	"encoding/json"

	"coin.local/coin-api/internal/componentpackage"
	"coin.local/coin-api/internal/manifest"
)

type compositionApplier func(parts *manifest.Composition, meta map[string]any, compName, compVersion string)

var compositionAppliers = map[string]compositionApplier{
	"agent":           applyAgentComposition,
	"gp-content":      applyGPContentComposition,
	"pipeline":        applyPipelineComposition,
	"branching-model": applyBranchingModelComposition,
}

func applyCompositionSlot(parts *manifest.Composition, typ, compName, compVersion string, meta map[string]any) {
	apply, ok := compositionAppliers[typ]
	if !ok {
		return
	}
	apply(parts, meta, compName, compVersion)
}

func applyAgentComposition(parts *manifest.Composition, meta map[string]any, _, _ string) {
	parts.AgentImage = metaString(meta, "image")
	parts.AgentDigest = metaString(meta, "digest")
}

func applyGPContentComposition(parts *manifest.Composition, meta map[string]any, compName, compVersion string) {
	parts.GPContentName = compName
	parts.GPContentVersion = compVersion
	_ = meta
}

func applyPipelineComposition(parts *manifest.Composition, meta map[string]any, _, compVersion string) {
	parts.PipelineVersion = compVersion
	_ = meta
}

func applyBranchingModelComposition(parts *manifest.Composition, meta map[string]any, compName, compVersion string) {
	parts.BranchingModelName = compName
	parts.BranchingModelVersion = compVersion
	_ = meta
}

func metaString(meta map[string]any, key string) string {
	if meta == nil {
		return ""
	}
	v, _ := meta[key].(string)
	return v
}

func contentBundleFromRawRef(meta gpContentMetadata, raw json.RawMessage) (manifest.ContentBundle, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return manifest.ContentBundle{}, nil
	}
	if isContentRefV2(raw) {
		return contentBundleFromV2Manifest(meta, raw)
	}
	var cref gpContentContentRef
	if err := json.Unmarshal(raw, &cref); err != nil {
		return manifest.ContentBundle{}, err
	}
	return contentBundleFromGPContent(meta, cref), nil
}

func contentBundleFromV2Manifest(meta gpContentMetadata, raw json.RawMessage) (manifest.ContentBundle, error) {
	var envelope struct {
		Manifest map[string]any `json:"manifest"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return manifest.ContentBundle{}, err
	}
	if envelope.Manifest == nil {
		return manifest.ContentBundle{}, nil
	}
	subsetRaw, err := json.Marshal(envelope.Manifest)
	if err != nil {
		return manifest.ContentBundle{}, err
	}
	var cref gpContentContentRef
	if err := json.Unmarshal(subsetRaw, &cref); err != nil {
		return manifest.ContentBundle{}, err
	}
	return contentBundleFromGPContent(meta, cref), nil
}

func isContentRefV2(raw json.RawMessage) bool {
	return componentpackage.IsContentRefV2Envelope(raw)
}
