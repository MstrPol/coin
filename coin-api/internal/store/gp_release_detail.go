package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type CompositionItem struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type GPReleaseDetail struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Status       string            `json:"status"`
	ManifestHash *string           `json:"manifestHash,omitempty"`
	ManifestURL  *string           `json:"manifestUrl,omitempty"`
	GitExportTag *string           `json:"gitExportTag,omitempty"`
	CreatedAt    time.Time         `json:"createdAt"`
	Composition  []CompositionItem `json:"composition"`
}

func (s *Store) GetGPReleaseDetail(ctx context.Context, name, version string) (GPReleaseDetail, error) {
	var detail GPReleaseDetail
	err := s.pool.QueryRow(ctx, `
		SELECT name, version, status, manifest_hash, manifest_url, git_export_tag, created_at
		FROM gp_releases
		WHERE name = $1 AND version = $2
	`, name, version).Scan(
		&detail.Name, &detail.Version, &detail.Status,
		&detail.ManifestHash, &detail.ManifestURL, &detail.GitExportTag, &detail.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return GPReleaseDetail{}, ErrNotFound
	}
	if err != nil {
		return GPReleaseDetail{}, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT gc.component_type, gc.component_name, gc.component_version
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2
		ORDER BY gc.component_type, gc.component_name
	`, name, version)
	if err != nil {
		return GPReleaseDetail{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item CompositionItem
		if err := rows.Scan(&item.Type, &item.Name, &item.Version); err != nil {
			return GPReleaseDetail{}, err
		}
		detail.Composition = append(detail.Composition, item)
	}
	if detail.Composition == nil {
		detail.Composition = []CompositionItem{}
	}
	return detail, rows.Err()
}
