package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/gpcontent"
	"coin.local/coin-api/internal/manifest"
)

var ErrGPReleasePipelineMissing = errors.New("gp release pipeline body not found")

type GPReleasePipelineBody struct {
	SchemaVersion int             `json:"schemaVersion"`
	Body          json.RawMessage `json:"body"`
	SHA256        string          `json:"sha256"`
}

func (s *Store) loadGPReleasePipelineRaw(ctx context.Context, gpName, gpVersion string) ([]byte, error) {
	var body []byte
	err := s.pool.QueryRow(ctx, `
		SELECT b.body
		FROM gp_release_pipeline_bodies b
		JOIN gp_releases gr ON gr.id = b.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2
	`, gpName, gpVersion).Scan(&body)
	if err == pgx.ErrNoRows {
		return nil, ErrGPReleasePipelineMissing
	}
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s *Store) loadContentBundleFromPipeline(ctx context.Context, gpName, gpVersion string) (manifest.ContentBundle, error) {
	raw, err := s.loadGPReleasePipelineRaw(ctx, gpName, gpVersion)
	if err != nil {
		return manifest.ContentBundle{}, err
	}
	doc, err := gpcontent.ParseDoc(raw)
	if err != nil {
		return manifest.ContentBundle{}, fmt.Errorf("pipeline body: %w", err)
	}
	doc.Name = gpName
	doc.Version = gpVersion
	bundle := gpcontent.ContentBundleFromInlineDoc(doc, gpName, gpVersion)
	if key := bundle.SchemaArtifactKey; key != "" {
		if sha, err := s.loadReleaseArtifactSHA256(ctx, gpName, gpVersion, key); err == nil {
			bundle.SchemaSHA256 = sha
		}
	}
	return bundle, nil
}

func (s *Store) loadReleaseArtifactSHA256(ctx context.Context, gpName, gpVersion, key string) (string, error) {
	var sha string
	err := s.pool.QueryRow(ctx, `
		SELECT gab.sha256
		FROM gp_artifact_bodies gab
		JOIN gp_releases gr ON gr.id = gab.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2 AND gab.artifact_key = $3
	`, gpName, gpVersion, key).Scan(&sha)
	return sha, err
}

func (s *Store) SaveGPReleasePipelineBody(ctx context.Context, gpName, gpVersion string, raw []byte) error {
	if err := s.requireGPReleaseDraft(ctx, gpName, gpVersion); err != nil {
		return err
	}
	doc, err := gpcontent.ParseDoc(raw)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidComposition, err)
	}
	issues, _ := gpcontent.ValidateDoc(doc, gpcontent.PreviewOptions{ComponentName: gpName})
	if len(issues) > 0 {
		return fmt.Errorf("%w: pipeline validation failed", ErrInvalidComposition)
	}
	canonical, err := gpcontent.PipelineBodyJSON(doc)
	if err != nil {
		return err
	}
	var releaseID int64
	err = s.pool.QueryRow(ctx, `
		SELECT id FROM gp_releases WHERE name = $1 AND version = $2
	`, gpName, gpVersion).Scan(&releaseID)
	if err != nil {
		return err
	}
	sum := pipelineBodySHA256(canonical)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO gp_release_pipeline_bodies (gp_release_id, schema_version, body, sha256, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (gp_release_id) DO UPDATE SET
			schema_version = EXCLUDED.schema_version,
			body = EXCLUDED.body,
			sha256 = EXCLUDED.sha256,
			updated_at = now()
	`, releaseID, doc.SchemaVersion, canonical, sum)
	return err
}

func (s *Store) GetGPReleasePipelineBody(ctx context.Context, gpName, gpVersion string) (GPReleasePipelineBody, error) {
	var out GPReleasePipelineBody
	err := s.pool.QueryRow(ctx, `
		SELECT b.schema_version, b.body, b.sha256
		FROM gp_release_pipeline_bodies b
		JOIN gp_releases gr ON gr.id = b.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2
	`, gpName, gpVersion).Scan(&out.SchemaVersion, &out.Body, &out.SHA256)
	if err == pgx.ErrNoRows {
		return GPReleasePipelineBody{}, ErrGPReleasePipelineMissing
	}
	return out, err
}

func (s *Store) seedGPReleasePipeline(ctx context.Context, releaseID int64, gpName, gpVersion string) error {
	raw, err := gpcontent.SeedPipelineBody(gpName, gpVersion)
	if err != nil {
		return err
	}
	doc, err := gpcontent.ParseDoc(raw)
	if err != nil {
		return err
	}
	sum := pipelineBodySHA256(raw)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO gp_release_pipeline_bodies (gp_release_id, schema_version, body, sha256, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (gp_release_id) DO NOTHING
	`, releaseID, doc.SchemaVersion, raw, sum)
	return err
}

func (s *Store) copyPipelineBetween(ctx context.Context, gpName, fromVersion, toVersion string) error {
	from, err := s.loadGPReleasePipelineRaw(ctx, gpName, fromVersion)
	if err != nil {
		return err
	}
	var toID int64
	if err := s.pool.QueryRow(ctx, `
		SELECT id FROM gp_releases WHERE name=$1 AND version=$2
	`, gpName, toVersion).Scan(&toID); err != nil {
		return err
	}
	doc, err := gpcontent.ParseDoc(from)
	if err != nil {
		return err
	}
	sum := pipelineBodySHA256(from)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO gp_release_pipeline_bodies (gp_release_id, schema_version, body, sha256, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (gp_release_id) DO NOTHING
	`, toID, doc.SchemaVersion, from, sum)
	return err
}

func (s *Store) seedGPReleaseSchemaArtifact(ctx context.Context, releaseID int64) error {
	body, err := gpcontent.ValidateSchemaBytes()
	if err != nil {
		return err
	}
	sum := pipelineBodySHA256(body)
	key := "schemas/config.v2.schema.json"
	_, err = s.pool.Exec(ctx, `
		INSERT INTO gp_artifact_bodies (gp_release_id, artifact_key, body, sha256)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (gp_release_id, artifact_key) DO NOTHING
	`, releaseID, key, body, sum)
	return err
}

func (s *Store) validateGPReleasePipeline(ctx context.Context, gpName, gpVersion string) error {
	_, err := s.loadContentBundleFromPipeline(ctx, gpName, gpVersion)
	return err
}

func (s *Store) requireGPReleaseDraft(ctx context.Context, gpName, gpVersion string) error {
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT status FROM gp_releases WHERE name = $1 AND version = $2
	`, gpName, gpVersion).Scan(&status)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "draft" {
		return ErrComponentVersionNotDraft
	}
	return nil
}

func pipelineBodySHA256(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
