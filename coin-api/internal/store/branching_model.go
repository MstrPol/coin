package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/manifest"
)

func (s *Store) branchingModelVersionFromComposition(ctx context.Context, gpName, gpVersion string) (string, string, error) {
	var name, ver string
	err := s.pool.QueryRow(ctx, `
		SELECT gc.component_name, gc.component_version
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2
		  AND gc.component_type = 'branching-model'
	`, gpName, gpVersion).Scan(&name, &ver)
	if err == pgx.ErrNoRows {
		return "", "", fmt.Errorf("branching-model not in composition for %s@%s", gpName, gpVersion)
	}
	return name, ver, err
}

func (s *Store) getBranchingModelRefs(ctx context.Context, name, version string, mode ComponentResolveMode) (json.RawMessage, error) {
	allowed := allowedComponentStatuses(mode)
	var crefRaw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(cv.content_ref, 'null'::jsonb)
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = 'branching-model' AND c.name = $1 AND cv.version = $2
		  AND cv.status::text = ANY($3)
	`, name, version, allowed).Scan(&crefRaw)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("branching-model/%s@%s not found or not visible for resolve mode %q", name, version, mode)
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(crefRaw), nil
}

func (s *Store) loadBranchingBundle(ctx context.Context, gpName, gpVersion string, mode ComponentResolveMode) (manifest.BranchingBundle, error) {
	bmName, bmVer, err := s.branchingModelVersionFromComposition(ctx, gpName, gpVersion)
	if err != nil {
		return manifest.BranchingBundle{}, err
	}
	crefRaw, err := s.getBranchingModelRefs(ctx, bmName, bmVer, mode)
	if err != nil {
		return manifest.BranchingBundle{}, err
	}
	rules, err := branchingRulesFromRawRef(crefRaw)
	if err != nil {
		return manifest.BranchingBundle{}, err
	}
	return manifest.BranchingBundle{
		Name:    bmName,
		Version: bmVer,
		Rules:   rules,
	}, nil
}

func branchingRulesFromRawRef(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	if isContentRefV2(raw) {
		return branchingRulesFromV2Manifest(raw)
	}
	return nil, fmt.Errorf("branching-model content_ref v2 required")
}

func branchingRulesFromV2Manifest(raw json.RawMessage) (map[string]any, error) {
	var envelope struct {
		Manifest map[string]any `json:"manifest"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}
	if envelope.Manifest == nil {
		return nil, nil
	}
	if branching, ok := envelope.Manifest["branching"].(map[string]any); ok {
		return cloneMapAny(branching), nil
	}
	return cloneMapAny(envelope.Manifest), nil
}

func (s *Store) loadBranchingBundleOptional(ctx context.Context, gpName, gpVersion string, mode ComponentResolveMode) (manifest.BranchingBundle, error) {
	bundle, err := s.loadBranchingBundle(ctx, gpName, gpVersion, mode)
	if err != nil && isBranchingNotPinned(err) {
		return manifest.BranchingBundle{}, nil
	}
	return bundle, err
}

func isBranchingNotPinned(err error) bool {
	return err != nil && strings.Contains(err.Error(), "branching-model not in composition")
}

func cloneMapAny(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
