package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"coin.local/coin-api/internal/manifest"
)

var ErrNotFound = errors.New("gp release not found")
var ErrUnsupportedGP = errors.New("unsupported golden path")
var ErrDuplicateGPProfile = errors.New("gp profile already exists")

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

type ReleaseRow struct {
	Name      string
	Version   string
	Parts     manifest.Composition
	Content   manifest.ContentBundle
	Branching manifest.BranchingBundle
}

func (s *Store) GetGPRelease(ctx context.Context, name, version string) (ReleaseRow, error) {
	var row ReleaseRow
	err := s.pool.QueryRow(ctx, `
		SELECT name, version FROM gp_releases
		WHERE name=$1 AND version=$2 AND status='published'
	`, name, version).Scan(&row.Name, &row.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return ReleaseRow{}, ErrNotFound
	}
	if err != nil {
		return ReleaseRow{}, fmt.Errorf("gp release: %w", err)
	}

	parts, err := s.loadComposition(ctx, name, version, ComponentResolveStable)
	if err != nil {
		return ReleaseRow{}, err
	}
	row.Parts = parts

	content, err := s.loadContentBundle(ctx, name, version, ComponentResolveStable)
	if err != nil {
		return ReleaseRow{}, err
	}
	row.Content = content

	branching, err := s.loadBranchingBundleOptional(ctx, name, version, ComponentResolveStable)
	if err != nil {
		return ReleaseRow{}, err
	}
	row.Branching = branching
	return row, nil
}

func (s *Store) loadComposition(ctx context.Context, name, version string, mode ComponentResolveMode) (manifest.Composition, error) {
	allowed := allowedComponentStatuses(mode)
	rows, err := s.pool.Query(ctx, `
		SELECT gc.component_type, gc.component_name, gc.component_version, cv.metadata, cv.status::text
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		JOIN component_versions cv ON cv.version = gc.component_version
		JOIN components c ON c.id = cv.component_id
			AND c.type = gc.component_type AND c.name = gc.component_name
		WHERE gr.name=$1 AND gr.version=$2
		  AND cv.status::text = ANY($3)
	`, name, version, allowed)
	if err != nil {
		return manifest.Composition{}, fmt.Errorf("composition: %w", err)
	}
	defer rows.Close()

	var parts manifest.Composition
	for rows.Next() {
		var typ, compName, compVersion, status string
		var metadata []byte
		if err := rows.Scan(&typ, &compName, &compVersion, &metadata, &status); err != nil {
			return manifest.Composition{}, err
		}
		if !componentStatusAllowed(status, mode) {
			return manifest.Composition{}, fmt.Errorf("component %s/%s@%s status %q not allowed for resolve mode %q", typ, compName, compVersion, status, mode)
		}
		var meta map[string]any
		_ = json.Unmarshal(metadata, &meta)
		applyCompositionSlot(&parts, typ, compName, compVersion, meta)
	}
	if err := rows.Err(); err != nil {
		return manifest.Composition{}, err
	}
	return parts, nil
}

func (s *Store) SaveManifestMeta(ctx context.Context, name, version, hash, url string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE gp_releases SET manifest_hash=$3, manifest_url=$4 WHERE name=$1 AND version=$2
	`, name, version, hash, url)
	return err
}

func (s *Store) ManifestURL(ctx context.Context, name, version string) (string, error) {
	var url *string
	err := s.pool.QueryRow(ctx, `
		SELECT manifest_url FROM gp_releases WHERE name=$1 AND version=$2
	`, name, version).Scan(&url)
	if err != nil {
		return "", err
	}
	if url == nil {
		return "", nil
	}
	return *url, nil
}
