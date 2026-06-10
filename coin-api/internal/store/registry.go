package store

import (
	"context"
	"time"
)

type DashboardStats struct {
	Projects     int `json:"projects"`
	GPReleases   int `json:"gpReleases"`
	BuildReports int `json:"buildReports"`
	GoldenPaths  int `json:"goldenPaths"`
}

type ProjectRow struct {
	Name       string    `json:"name"`
	GoldenPath string    `json:"goldenPath"`
	Version    string    `json:"version"`
	CanaryMode string    `json:"canaryMode"`
	GitURL     *string   `json:"gitUrl,omitempty"`
	LastSeenAt time.Time `json:"lastSeenAt"`
}

type GPReleaseListItem struct {
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	Status       string    `json:"status"`
	ManifestHash *string   `json:"manifestHash,omitempty"`
	ManifestURL  *string   `json:"manifestUrl,omitempty"`
	GitExportTag *string   `json:"gitExportTag,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (s *Store) DashboardStats(ctx context.Context) (DashboardStats, error) {
	var stats DashboardStats
	err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::int FROM projects),
			(SELECT COUNT(*)::int FROM gp_releases WHERE status = 'published'),
			(SELECT COUNT(*)::int FROM build_reports),
			(SELECT COUNT(*)::int FROM gp_profiles)
	`).Scan(&stats.Projects, &stats.GPReleases, &stats.BuildReports, &stats.GoldenPaths)
	return stats, err
}

func (s *Store) ListProjects(ctx context.Context, goldenPath, version string) ([]ProjectRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.name, pb.gp_name, pb.gp_version, p.canary_mode::text, pb.git_url, pb.last_seen_at
		FROM projects p
		JOIN LATERAL (
			SELECT gp_name, gp_version, git_url, last_seen_at
			FROM project_bindings pb2
			WHERE pb2.project_id = p.id
			ORDER BY pb2.last_seen_at DESC
			LIMIT 1
		) pb ON true
		WHERE ($1 = '' OR pb.gp_name = $1)
		  AND ($2 = '' OR pb.gp_version = $2)
		ORDER BY p.name
	`, goldenPath, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ProjectRow
	for rows.Next() {
		var row ProjectRow
		if err := rows.Scan(&row.Name, &row.GoldenPath, &row.Version, &row.CanaryMode, &row.GitURL, &row.LastSeenAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Store) ListGPReleases(ctx context.Context, name string, includeDrafts bool) ([]GPReleaseListItem, error) {
	query := `
		SELECT name, version, status, manifest_hash, manifest_url, git_export_tag, created_at
		FROM gp_releases
		WHERE 1=1
	`
	if !includeDrafts {
		query += ` AND status = 'published'`
	}
	args := []any{}
	if name != "" {
		query += ` AND name = $1`
		args = append(args, name)
	}
	query += ` ORDER BY name, created_at DESC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []GPReleaseListItem
	for rows.Next() {
		var row GPReleaseListItem
		if err := rows.Scan(
			&row.Name, &row.Version, &row.Status,
			&row.ManifestHash, &row.ManifestURL, &row.GitExportTag, &row.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
