package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/pin"
)

func (s *Store) ListPublishedGPVersions(ctx context.Context, name string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT version FROM gp_releases
		WHERE name=$1 AND status='published'
		  AND version NOT LIKE '%-snapshot.%'
		ORDER BY version
	`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (s *Store) GetGPReleaseForResolve(ctx context.Context, name, version string, opts GPResolveOptions) (ReleaseRow, error) {
	statusFilter := "status='published'"
	if opts.AllowDraftGP {
		statusFilter = "status IN ('published', 'draft')"
	}
	query := fmt.Sprintf(`
		SELECT name, version, image_registry_prefix, build_cache_enabled, artifact_repository_base FROM gp_releases
		WHERE name=$1 AND version=$2 AND %s
	`, statusFilter)

	var row ReleaseRow
	err := s.pool.QueryRow(ctx, query, name, version).Scan(
		&row.Name,
		&row.Version,
		&row.Destinations.ImageRegistryPrefix,
		&row.Destinations.BuildCacheEnabled,
		&row.Destinations.ArtifactRepositoryBase,
	)
	if err == pgx.ErrNoRows {
		return ReleaseRow{}, ErrNotFound
	}
	if err != nil {
		return ReleaseRow{}, fmt.Errorf("gp release: %w", err)
	}

	mode := opts.ComponentMode
	if mode == "" {
		mode = ComponentResolveStable
	}

	parts, err := s.loadComposition(ctx, name, version, mode)
	if err != nil {
		return ReleaseRow{}, err
	}
	row.Parts = parts

	content, err := s.loadContentBundle(ctx, name, version, mode)
	if err != nil {
		return ReleaseRow{}, err
	}
	row.Content = content

	branching, err := s.loadBranchingBundleOptional(ctx, name, version, mode)
	if err != nil {
		return ReleaseRow{}, err
	}
	row.Branching = branching
	return row, nil
}

func (s *Store) CreateDraftGPRelease(ctx context.Context, in PublishGPReleaseInput) (GPReleaseRow, error) {
	if in.Name == "" || in.Version == "" {
		return GPReleaseRow{}, fmt.Errorf("name and version are required")
	}
	if err := validateReleaseDestinations(in.Destinations); err != nil {
		return GPReleaseRow{}, err
	}

	prep, err := s.prepareGPRelease(ctx, in)
	if err != nil {
		return GPReleaseRow{}, err
	}

	rules, err := s.loadCompatibilityRules(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	if err := validateGPReleaseComposition(s, ctx, prep, rules, componentResolveModeForGPDraftEdit); err != nil {
		return GPReleaseRow{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	defer tx.Rollback(ctx)

	var releaseID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO gp_releases (name, version, status, image_registry_prefix, build_cache_enabled, artifact_repository_base)
		VALUES ($1, $2, 'draft', $3, $4, $5)
		RETURNING id
	`, in.Name, in.Version, in.Destinations.ImageRegistryPrefix, in.Destinations.BuildCacheEnabled, in.Destinations.ArtifactRepositoryBase).Scan(&releaseID)
	if err != nil {
		if isUniqueViolation(err) {
			return GPReleaseRow{}, ErrDuplicateGPRelease
		}
		return GPReleaseRow{}, fmt.Errorf("draft insert: %w", err)
	}

	if err := insertGPComposition(ctx, tx, releaseID, prep.storeSlots, in.Composition); err != nil {
		return GPReleaseRow{}, err
	}

	entityKey := fmt.Sprintf("%s@%s", in.Name, in.Version)
	auditPayload, _ := json.Marshal(map[string]any{
		"name":        in.Name,
		"version":     in.Version,
		"composition": in.Composition,
		"status":      "draft",
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('create_gp_draft', 'gp_release', $1, $2, $3)
	`, entityKey, nullIfEmpty(in.Actor), auditPayload)
	if err != nil {
		return GPReleaseRow{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return GPReleaseRow{}, err
	}

	// Seed pipeline body from latest published release or embedded default.
	seeded := false
	if latest, err := s.latestPublishedVersion(ctx, in.Name); err == nil && latest != "" {
		if err := s.copyPipelineBetween(ctx, in.Name, latest, in.Version); err == nil {
			seeded = true
		}
	}
	if !seeded {
		_ = s.seedGPReleasePipeline(ctx, releaseID, in.Name, in.Version)
	}
	_ = s.seedGPReleaseSchemaArtifact(ctx, releaseID)

	return GPReleaseRow{Name: in.Name, Version: in.Version, Status: "draft"}, nil
}

func (s *Store) latestPublishedVersion(ctx context.Context, gpName string) (string, error) {
	var version string
	err := s.pool.QueryRow(ctx, `
		SELECT version FROM gp_releases
		WHERE name=$1 AND status='published'
		ORDER BY created_at DESC LIMIT 1
	`, gpName).Scan(&version)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return version, err
}

func (s *Store) copyArtifactsBetween(ctx context.Context, gpName, fromVersion, toVersion string) error {
	from, err := s.ListArtifactBodies(ctx, gpName, fromVersion)
	if err != nil {
		return err
	}
	var toID int64
	err = s.pool.QueryRow(ctx, `
		SELECT id FROM gp_releases WHERE name=$1 AND version=$2
	`, gpName, toVersion).Scan(&toID)
	if err != nil {
		return err
	}
	for _, a := range from {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO gp_artifact_bodies (gp_release_id, artifact_key, body, sha256)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (gp_release_id, artifact_key) DO NOTHING
		`, toID, a.Key, a.Body, a.SHA256)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) PromoteDraftToPublished(ctx context.Context, name, version, actor string) (GPReleaseRow, error) {
	publishedVersion := pin.StripSnapshotVersion(version)
	if publishedVersion == "" {
		return GPReleaseRow{}, fmt.Errorf("invalid draft version %q", version)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	defer tx.Rollback(ctx)

	var releaseID int64
	var status string
	err = tx.QueryRow(ctx, `
		SELECT id, status FROM gp_releases WHERE name=$1 AND version=$2 FOR UPDATE
	`, name, version).Scan(&releaseID, &status)
	if err == pgx.ErrNoRows {
		return GPReleaseRow{}, ErrNotFound
	}
	if err != nil {
		return GPReleaseRow{}, err
	}
	if status != "draft" {
		return GPReleaseRow{}, fmt.Errorf("release is not draft (status=%s)", status)
	}

	if blockers, err := s.validateGPReleasePromoteReady(ctx, name, version); err != nil {
		if errors.Is(err, ErrGPCompositionHasDraftPins) {
			return GPReleaseRow{}, fmt.Errorf("%w: %v", ErrGPCompositionHasDraftPins, blockers)
		}
		return GPReleaseRow{}, err
	}

	var exists int
	err = tx.QueryRow(ctx, `
		SELECT 1 FROM gp_releases
		WHERE name=$1 AND version=$2 AND status='published' AND id <> $3
	`, name, publishedVersion, releaseID).Scan(&exists)
	if err == nil {
		return GPReleaseRow{}, ErrDuplicateGPRelease
	}
	if err != nil && err != pgx.ErrNoRows {
		return GPReleaseRow{}, err
	}

	if publishedVersion != version {
		_, err = tx.Exec(ctx, `
			UPDATE gp_releases SET version=$1, status='published' WHERE id=$2
		`, publishedVersion, releaseID)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE gp_releases SET status='published' WHERE id=$1
		`, releaseID)
	}
	if err != nil {
		if isUniqueViolation(err) {
			return GPReleaseRow{}, ErrDuplicateGPRelease
		}
		return GPReleaseRow{}, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO catalog_policy (gp_name, latest, minimum, deprecated)
		VALUES ($1, $2, $2, '[]'::jsonb)
		ON CONFLICT (gp_name) DO UPDATE SET latest = EXCLUDED.latest
	`, name, publishedVersion)
	if err != nil {
		return GPReleaseRow{}, err
	}

	auditPayload, _ := json.Marshal(map[string]any{
		"status":           "published",
		"draftVersion":     version,
		"publishedVersion": publishedVersion,
	})
	entityKey := fmt.Sprintf("%s@%s", name, publishedVersion)
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('publish_gp_draft', 'gp_release', $1, $2, $3)
	`, entityKey, nullIfEmpty(actor), auditPayload)
	if err != nil {
		return GPReleaseRow{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return GPReleaseRow{}, err
	}
	return GPReleaseRow{Name: name, Version: publishedVersion, Status: "published"}, nil
}
