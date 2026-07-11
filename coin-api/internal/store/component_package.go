package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ComponentArtifactBody struct {
	Key    string
	Body   []byte
	SHA256 string
}

func (s *Store) ListComponentArtifactBodies(ctx context.Context, typ, name, version string) ([]ComponentArtifactBody, error) {
	if typ == "" || name == "" || version == "" {
		return nil, fmt.Errorf("type, name and version are required")
	}
	if err := s.requireComponentVersionDraft(ctx, typ, name, version); err != nil {
		return nil, err
	}
	var versionID int64
	err := s.pool.QueryRow(ctx, `
		SELECT cv.id FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version).Scan(&versionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT artifact_key, body, sha256
		FROM component_artifact_bodies
		WHERE component_version_id = $1
		ORDER BY artifact_key
	`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ComponentArtifactBody
	for rows.Next() {
		var row ComponentArtifactBody
		if err := rows.Scan(&row.Key, &row.Body, &row.SHA256); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Store) ListComponentArtifactBodiesForVersion(ctx context.Context, typ, name, version string) ([]ComponentArtifactBody, error) {
	if typ == "" || name == "" || version == "" {
		return nil, fmt.Errorf("type, name and version are required")
	}
	var versionID int64
	err := s.pool.QueryRow(ctx, `
		SELECT cv.id FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version).Scan(&versionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT artifact_key, body, sha256
		FROM component_artifact_bodies
		WHERE component_version_id = $1
		ORDER BY artifact_key
	`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ComponentArtifactBody
	for rows.Next() {
		var row ComponentArtifactBody
		if err := rows.Scan(&row.Key, &row.Body, &row.SHA256); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Store) GetComponentArtifactBody(ctx context.Context, typ, name, version, key string) ([]byte, string, error) {
	if typ == "" || name == "" || version == "" || key == "" {
		return nil, "", fmt.Errorf("type, name, version and artifact key are required")
	}
	if err := s.requireComponentVersionDraft(ctx, typ, name, version); err != nil {
		return nil, "", err
	}
	var versionID int64
	err := s.pool.QueryRow(ctx, `
		SELECT cv.id FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version).Scan(&versionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	return s.loadComponentArtifactBody(ctx, versionID, key)
}

func (s *Store) ListComponentArtifactMeta(ctx context.Context, typ, name, version string) ([]ComponentArtifactMeta, error) {
	if typ == "" || name == "" || version == "" {
		return nil, fmt.Errorf("type, name and version are required")
	}
	if err := s.requireComponentVersionDraft(ctx, typ, name, version); err != nil {
		return nil, err
	}
	var versionID int64
	err := s.pool.QueryRow(ctx, `
		SELECT cv.id FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version).Scan(&versionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT artifact_key, sha256, length(body)
		FROM component_artifact_bodies
		WHERE component_version_id = $1
		ORDER BY artifact_key
	`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ComponentArtifactMeta
	for rows.Next() {
		var row ComponentArtifactMeta
		if err := rows.Scan(&row.Key, &row.SHA256, &row.Size); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

type ComponentArtifactMeta struct {
	Key    string `json:"key"`
	SHA256 string `json:"sha256"`
	Size   int    `json:"size"`
}
