package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ArtifactMeta struct {
	Key    string `json:"key"`
	SHA256 string `json:"sha256"`
	Size   int    `json:"size"`
}

func (s *Store) ListArtifactMeta(ctx context.Context, gpName, gpVersion string) ([]ArtifactMeta, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ab.artifact_key, ab.sha256, length(ab.body)
		FROM gp_artifact_bodies ab
		JOIN gp_releases gr ON gr.id = ab.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2
		ORDER BY ab.artifact_key
	`, gpName, gpVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ArtifactMeta
	for rows.Next() {
		var item ArtifactMeta
		if err := rows.Scan(&item.Key, &item.SHA256, &item.Size); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetArtifactBody(ctx context.Context, gpName, gpVersion, key string) (ArtifactBody, error) {
	var item ArtifactBody
	err := s.pool.QueryRow(ctx, `
		SELECT ab.artifact_key, ab.body, ab.sha256
		FROM gp_artifact_bodies ab
		JOIN gp_releases gr ON gr.id = ab.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2 AND ab.artifact_key = $3
	`, gpName, gpVersion, key).Scan(&item.Key, &item.Body, &item.SHA256)
	if errors.Is(err, pgx.ErrNoRows) {
		return ArtifactBody{}, ErrNotFound
	}
	if err != nil {
		return ArtifactBody{}, err
	}
	return item, nil
}

func (s *Store) UpsertArtifactBody(ctx context.Context, gpName, gpVersion, key string, body []byte) error {
	var releaseID int64
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT id, status FROM gp_releases WHERE name=$1 AND version=$2
	`, gpName, gpVersion).Scan(&releaseID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "draft" {
		return fmt.Errorf("artifact edit allowed only for draft releases")
	}

	sum := sha256.Sum256(body)
	hash := "sha256:" + hex.EncodeToString(sum[:])

	_, err = s.pool.Exec(ctx, `
		INSERT INTO gp_artifact_bodies (gp_release_id, artifact_key, body, sha256)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (gp_release_id, artifact_key)
		DO UPDATE SET body = EXCLUDED.body, sha256 = EXCLUDED.sha256
	`, releaseID, key, body, hash)
	return err
}
