package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/compatibility"
	"coin.local/coin-api/internal/manifest"
)

var (
	ErrDuplicateGPRelease   = errors.New("gp release already exists")
	ErrComponentNotFound    = errors.New("component version not found")
	ErrInvalidComposition   = errors.New("invalid composition")
)

type PublishGPReleaseInput struct {
	Name        string
	Version     string
	Composition map[string]string
	Actor       string
}

type GPReleaseRow struct {
	ID           int64
	Name         string
	Version      string
	GitExportTag string
	Status       string
}

func (s *Store) PublishGPRelease(ctx context.Context, in PublishGPReleaseInput) (GPReleaseRow, error) {
	if in.Name == "" || in.Version == "" {
		return GPReleaseRow{}, fmt.Errorf("name and version are required")
	}

	slots, err := s.profileSlots(ctx, in.Name)
	if err != nil {
		return GPReleaseRow{}, err
	}

	rules, err := s.loadCompatibilityRules(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	if err := compatibility.Validate(slots, in.Composition, rules); err != nil {
		return GPReleaseRow{}, fmt.Errorf("%w: %v", ErrInvalidComposition, err)
	}

	for _, slot := range slots {
		ver := in.Composition[slot.Key]
		ok, err := s.componentVersionPublished(ctx, slot.Type, slot.Name, ver)
		if err != nil {
			return GPReleaseRow{}, err
		}
		if !ok {
			return GPReleaseRow{}, fmt.Errorf("%w: %s/%s@%s", ErrComponentNotFound, slot.Type, slot.Name, ver)
		}
	}

	var releaseID int64
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO gp_releases (name, version, status)
		VALUES ($1, $2, 'published')
		RETURNING id
	`, in.Name, in.Version).Scan(&releaseID)
	if err != nil {
		if isUniqueViolation(err) {
			return GPReleaseRow{}, ErrDuplicateGPRelease
		}
		return GPReleaseRow{}, fmt.Errorf("gp release insert: %w", err)
	}

	for _, slot := range slots {
		_, err = tx.Exec(ctx, `
			INSERT INTO gp_composition (gp_release_id, component_type, component_name, component_version)
			VALUES ($1, $2, $3, $4)
		`, releaseID, slot.Type, slot.Name, in.Composition[slot.Key])
		if err != nil {
			return GPReleaseRow{}, fmt.Errorf("composition insert: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO catalog_policy (gp_name, latest, minimum, deprecated)
		VALUES ($1, $2, $2, '[]'::jsonb)
		ON CONFLICT (gp_name) DO UPDATE SET latest = EXCLUDED.latest
	`, in.Name, in.Version)
	if err != nil {
		return GPReleaseRow{}, fmt.Errorf("catalog policy: %w", err)
	}

	entityKey := fmt.Sprintf("%s@%s", in.Name, in.Version)
	auditPayload, _ := json.Marshal(map[string]any{
		"name":        in.Name,
		"version":     in.Version,
		"composition": in.Composition,
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('publish_gp_release', 'gp_release', $1, $2, $3)
	`, entityKey, nullIfEmpty(in.Actor), auditPayload)
	if err != nil {
		return GPReleaseRow{}, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return GPReleaseRow{}, err
	}

	return GPReleaseRow{
		ID:      releaseID,
		Name:    in.Name,
		Version: in.Version,
		Status:  "published",
	}, nil
}

func (s *Store) componentVersionPublished(ctx context.Context, typ, name, version string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM component_versions cv
			JOIN components c ON c.id = cv.component_id
			WHERE c.type = $1 AND c.name = $2 AND cv.version = $3 AND cv.status = 'published'
		)
	`, typ, name, version).Scan(&exists)
	return exists, err
}

func (s *Store) loadCompatibilityRules(ctx context.Context) ([]compatibility.Rule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT source_type, source_name, source_version_prefix, requirements
		FROM component_compatibility
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []compatibility.Rule
	for rows.Next() {
		var rule compatibility.Rule
		var raw []byte
		if err := rows.Scan(&rule.SourceType, &rule.SourceName, &rule.VersionPrefix, &raw); err != nil {
			return nil, err
		}
		reqs, err := compatibility.ParseRequirements(raw)
		if err != nil {
			return nil, err
		}
		rule.Requirements = reqs
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (s *Store) getCompositionContentRef(ctx context.Context, gpName, gpVersion, typ, compName string) (json.RawMessage, error) {
	var ref []byte
	err := s.pool.QueryRow(ctx, `
		SELECT cv.content_ref
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		JOIN components c ON c.type = gc.component_type AND c.name = gc.component_name
		JOIN component_versions cv ON cv.component_id = c.id AND cv.version = gc.component_version
		WHERE gr.name = $1 AND gr.version = $2
		  AND gc.component_type = $3 AND gc.component_name = $4
	`, gpName, gpVersion, typ, compName).Scan(&ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("content ref not found for %s/%s in %s@%s", typ, compName, gpName, gpVersion)
	}
	if err != nil {
		return nil, err
	}
	if len(ref) == 0 {
		return nil, fmt.Errorf("empty content_ref for %s/%s", typ, compName)
	}
	return ref, nil
}

type refDoc struct {
	ArtifactKey string `json:"artifactKey"`
	SHA256      string `json:"sha256"`
}

type stageRefDoc struct {
	Name        string `json:"name"`
	When        string `json:"when"`
	ArtifactKey string `json:"artifactKey"`
	SHA256      string `json:"sha256"`
}

type pipelineRefDoc struct {
	Stages []stageRefDoc `json:"stages"`
}

func (s *Store) loadContentBundle(ctx context.Context, gpName, gpVersion string) (manifest.ContentBundle, error) {
	pipelineRaw, err := s.getCompositionContentRef(ctx, gpName, gpVersion, "pipeline", "go-build")
	if err != nil {
		return manifest.ContentBundle{}, err
	}
	validateRaw, err := s.getCompositionContentRef(ctx, gpName, gpVersion, "validate", "config")
	if err != nil {
		return manifest.ContentBundle{}, err
	}
	dockerfileRaw, err := s.getCompositionContentRef(ctx, gpName, gpVersion, "dockerfile", "go-runtime")
	if err != nil {
		return manifest.ContentBundle{}, err
	}
	orchRaw, err := s.getCompositionContentRef(ctx, gpName, gpVersion, "orchestration", "coin-pipeline")
	if err != nil {
		return manifest.ContentBundle{}, err
	}

	var pipeline pipelineRefDoc
	if err := json.Unmarshal(pipelineRaw, &pipeline); err != nil {
		return manifest.ContentBundle{}, fmt.Errorf("pipeline content_ref: %w", err)
	}
	var validate refDoc
	if err := json.Unmarshal(validateRaw, &validate); err != nil {
		return manifest.ContentBundle{}, fmt.Errorf("validate content_ref: %w", err)
	}
	var dockerfile refDoc
	if err := json.Unmarshal(dockerfileRaw, &dockerfile); err != nil {
		return manifest.ContentBundle{}, fmt.Errorf("dockerfile content_ref: %w", err)
	}
	var orch refDoc
	if err := json.Unmarshal(orchRaw, &orch); err != nil {
		return manifest.ContentBundle{}, fmt.Errorf("orchestration content_ref: %w", err)
	}

	stages := make([]manifest.StageScript, 0, len(pipeline.Stages))
	for _, st := range pipeline.Stages {
		stages = append(stages, manifest.StageScript{
			Name:        st.Name,
			When:        st.When,
			ArtifactKey: st.ArtifactKey,
			SHA256:      st.SHA256,
		})
	}
	if len(stages) == 0 {
		return manifest.ContentBundle{}, fmt.Errorf("pipeline content_ref missing stages for %s@%s", gpName, gpVersion)
	}

	return manifest.ContentBundle{
		SchemaArtifactKey:        validate.ArtifactKey,
		SchemaSHA256:             validate.SHA256,
		DockerfileArtifactKey:    dockerfile.ArtifactKey,
		DockerfileSHA256:         dockerfile.SHA256,
		OrchestrationArtifactKey: orch.ArtifactKey,
		OrchestrationSHA256:      orch.SHA256,
		Stages:                   stages,
	}, nil
}
