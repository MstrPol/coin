package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type ComponentListItem struct {
	Type            string    `json:"type"`
	Name            string    `json:"name"`
	LatestVersion   string    `json:"latestVersion"`
	VersionCount    int       `json:"versionCount"`
	LatestCreatedAt time.Time `json:"latestCreatedAt"`
}

type ComponentVersionListItem struct {
	Version   string    `json:"version"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *Store) ListComponents(ctx context.Context) ([]ComponentListItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.type, c.name,
		       COUNT(cv.id) FILTER (WHERE cv.status = 'published')::int,
		       COALESCE(
		           MAX(cv.created_at) FILTER (WHERE cv.status = 'published'),
		           to_timestamp(0)
		       ),
		       (
		           SELECT cv2.version
		           FROM component_versions cv2
		           WHERE cv2.component_id = c.id AND cv2.status = 'published'
		           ORDER BY cv2.created_at DESC
		           LIMIT 1
		       )
		FROM components c
		LEFT JOIN component_versions cv ON cv.component_id = c.id
		GROUP BY c.id, c.type, c.name
		ORDER BY c.type, c.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ComponentListItem
	for rows.Next() {
		var row ComponentListItem
		var latestVersion *string
		if err := rows.Scan(&row.Type, &row.Name, &row.VersionCount, &row.LatestCreatedAt, &latestVersion); err != nil {
			return nil, err
		}
		if latestVersion != nil {
			row.LatestVersion = *latestVersion
		}
		out = append(out, row)
	}
	if out == nil {
		out = []ComponentListItem{}
	}
	return out, rows.Err()
}

func (s *Store) ListComponentVersions(ctx context.Context, typ, name string) ([]ComponentVersionListItem, error) {
	ok, err := s.componentExists(ctx, typ, name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}

	rows, err := s.pool.Query(ctx, `
		SELECT cv.version, cv.status, cv.created_at
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2
		ORDER BY cv.created_at DESC
	`, typ, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ComponentVersionListItem
	for rows.Next() {
		var row ComponentVersionListItem
		if err := rows.Scan(&row.Version, &row.Status, &row.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if out == nil {
		out = []ComponentVersionListItem{}
	}
	return out, rows.Err()
}

func (s *Store) componentExists(ctx context.Context, typ, name string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM components WHERE type = $1 AND name = $2)
	`, typ, name).Scan(&exists)
	return exists, err
}

type ComponentDetail struct {
	Type            string                  `json:"type"`
	Name            string                  `json:"name"`
	LatestVersion   string                  `json:"latestVersion"`
	VersionCount    int                     `json:"versionCount"`
	LatestCreatedAt time.Time               `json:"latestCreatedAt"`
	GPUsage         []ComponentGPUsageEntry `json:"gpUsage"`
}

type ComponentGPUsageEntry struct {
	GPName  string `json:"gpName"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

type ComponentVersionDetail struct {
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Version    string          `json:"version"`
	Status     string          `json:"status"`
	Metadata   json.RawMessage `json:"metadata"`
	ContentRef json.RawMessage `json:"contentRef,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}

func (s *Store) GetComponentDetail(ctx context.Context, typ, name string) (ComponentDetail, error) {
	ok, err := s.componentExists(ctx, typ, name)
	if err != nil {
		return ComponentDetail{}, err
	}
	if !ok {
		return ComponentDetail{}, ErrNotFound
	}

	var detail ComponentDetail
	detail.Type = typ
	detail.Name = name
	var latestVersion *string
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(cv.id) FILTER (WHERE cv.status = 'published')::int,
		       COALESCE(MAX(cv.created_at) FILTER (WHERE cv.status = 'published'), to_timestamp(0)),
		       (
		           SELECT cv2.version
		           FROM component_versions cv2
		           WHERE cv2.component_id = c.id AND cv2.status = 'published'
		           ORDER BY cv2.created_at DESC
		           LIMIT 1
		       )
		FROM components c
		LEFT JOIN component_versions cv ON cv.component_id = c.id
		WHERE c.type = $1 AND c.name = $2
		GROUP BY c.id
	`, typ, name).Scan(&detail.VersionCount, &detail.LatestCreatedAt, &latestVersion)
	if latestVersion != nil {
		detail.LatestVersion = *latestVersion
	}
	if err != nil {
		return ComponentDetail{}, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT gr.name, gr.version, gr.status::text
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		WHERE gc.component_type = $1 AND gc.component_name = $2
		ORDER BY gr.name, gr.created_at DESC
	`, typ, name)
	if err != nil {
		return ComponentDetail{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry ComponentGPUsageEntry
		if err := rows.Scan(&entry.GPName, &entry.Version, &entry.Status); err != nil {
			return ComponentDetail{}, err
		}
		detail.GPUsage = append(detail.GPUsage, entry)
	}
	if detail.GPUsage == nil {
		detail.GPUsage = []ComponentGPUsageEntry{}
	}
	return detail, rows.Err()
}

func (s *Store) GetComponentVersionDetail(ctx context.Context, typ, name, version string) (ComponentVersionDetail, error) {
	ok, err := s.componentExists(ctx, typ, name)
	if err != nil {
		return ComponentVersionDetail{}, err
	}
	if !ok {
		return ComponentVersionDetail{}, ErrNotFound
	}

	var detail ComponentVersionDetail
	var contentRef []byte
	err = s.pool.QueryRow(ctx, `
		SELECT cv.version, cv.status::text, cv.metadata, cv.content_ref, cv.created_at
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version).Scan(
		&detail.Version, &detail.Status, &detail.Metadata, &contentRef, &detail.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ComponentVersionDetail{}, ErrNotFound
		}
		return ComponentVersionDetail{}, err
	}
	detail.Type = typ
	detail.Name = name
	if len(contentRef) > 0 {
		detail.ContentRef = contentRef
	}
	return detail, nil
}
