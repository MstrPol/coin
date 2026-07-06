package store

import (
	"context"
	"encoding/json"
	"errors"

	"coin.local/coin-api/internal/manifest"
)

var ErrGPCompositionHasDraftPins = errors.New("gp composition has draft component pins")

type CompositionPinBlocker struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

func (s *Store) validateGPCompositionPublishedPins(ctx context.Context, gpName, gpVersion string) ([]CompositionPinBlocker, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT gc.component_type, gc.component_name, gc.component_version, cv.status::text
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		JOIN component_versions cv ON cv.version = gc.component_version
		JOIN components c ON c.id = cv.component_id
			AND c.type = gc.component_type AND c.name = gc.component_name
		WHERE gr.name = $1 AND gr.version = $2
	`, gpName, gpVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blockers []CompositionPinBlocker
	for rows.Next() {
		var typ, name, version, status string
		if err := rows.Scan(&typ, &name, &version, &status); err != nil {
			return nil, err
		}
		if status != "published" {
			blockers = append(blockers, CompositionPinBlocker{
				Type: typ, Name: name, Version: version, Status: status,
			})
		}
	}
	return blockers, rows.Err()
}

func (s *Store) loadGPCompositionInput(ctx context.Context, gpName, gpVersion string) (PublishGPReleaseInput, error) {
	var dest manifest.Destinations
	if err := s.pool.QueryRow(ctx, `
		SELECT image_registry_prefix, build_cache_enabled, artifact_repository_base
		FROM gp_releases
		WHERE name = $1 AND version = $2
	`, gpName, gpVersion).Scan(&dest.ImageRegistryPrefix, &dest.BuildCacheEnabled, &dest.ArtifactRepositoryBase); err != nil {
		return PublishGPReleaseInput{}, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT gc.component_type, gc.component_name, gc.component_version
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2
	`, gpName, gpVersion)
	if err != nil {
		return PublishGPReleaseInput{}, err
	}
	defer rows.Close()

	composition := make(map[string]string)
	var agentStack, branching string
	for rows.Next() {
		var typ, name, version string
		if err := rows.Scan(&typ, &name, &version); err != nil {
			return PublishGPReleaseInput{}, err
		}
		switch typ {
		case "agent":
			composition["agent"] = version
			agentStack = name
		case "branching-model":
			composition["branching-model"] = version
			branching = name
		}
	}
	if err := rows.Err(); err != nil {
		return PublishGPReleaseInput{}, err
	}
	return PublishGPReleaseInput{
		Name:               gpName,
		Version:            gpVersion,
		Destinations:       dest,
		Composition:        composition,
		AgentStackName:     agentStack,
		BranchingModelName: branching,
	}, nil
}

func (s *Store) validateGPReleasePromoteReady(ctx context.Context, gpName, gpVersion string) ([]CompositionPinBlocker, error) {
	blockers, err := s.validateGPCompositionPublishedPins(ctx, gpName, gpVersion)
	if err != nil {
		return nil, err
	}
	if len(blockers) > 0 {
		return blockers, ErrGPCompositionHasDraftPins
	}

	in, err := s.loadGPCompositionInput(ctx, gpName, gpVersion)
	if err != nil {
		return nil, err
	}
	prep, err := s.prepareGPRelease(ctx, in)
	if err != nil {
		return nil, err
	}
	rules, err := s.loadCompatibilityRules(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateGPReleaseComposition(s, ctx, prep, rules, componentResolveModeForGPPromote); err != nil {
		return nil, err
	}
	if err := s.validateGPReleasePipeline(ctx, gpName, gpVersion); err != nil {
		return nil, err
	}
	return nil, nil
}

// ValidateGPReleasePromoteBlockers returns draft pins blocking GP promote (for API error payload).
func (s *Store) ValidateGPReleasePromoteBlockers(ctx context.Context, gpName, gpVersion string) ([]CompositionPinBlocker, error) {
	return s.validateGPCompositionPublishedPins(ctx, gpName, gpVersion)
}

func encodePromoteBlockers(blockers []CompositionPinBlocker) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{"blockingPins": blockers})
	return raw
}
