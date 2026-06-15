package store

import (
	"context"
	"time"
)

type DashboardStats struct {
	Projects      int `json:"projects"`
	StaleProjects int `json:"staleProjects"`
	GPReleases    int `json:"gpReleases"`
	BuildReports  int `json:"buildReports"`
	GoldenPaths   int `json:"goldenPaths"`
}

type ProjectRow struct {
	Name         string     `json:"name"`
	GroupID      string     `json:"groupId,omitempty"`
	ArtifactID   string     `json:"artifactId,omitempty"`
	GitRepoName  string     `json:"gitRepoName,omitempty"`
	GitRepoURL   string     `json:"gitRepoUrl,omitempty"`
	GoldenPath   string     `json:"goldenPath"`
	Version      string     `json:"version"`
	CanaryMode   string     `json:"canaryMode"`
	Branch       string     `json:"branch,omitempty"`
	LastBuildAt  *time.Time `json:"lastBuildAt,omitempty"`
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
			(SELECT COUNT(*)::int FROM projects p
			 WHERE NOT EXISTS (
			   SELECT 1 FROM build_reports br
			   WHERE br.project_id = p.id
			     AND br.reported_at > now() - interval '90 days'
			 )),
			(SELECT COUNT(*)::int FROM gp_releases WHERE status = 'published'),
			(SELECT COUNT(*)::int FROM build_reports),
			(SELECT COUNT(*)::int FROM gp_profiles)
	`).Scan(&stats.Projects, &stats.StaleProjects, &stats.GPReleases, &stats.BuildReports, &stats.GoldenPaths)
	return stats, err
}

func (s *Store) ListProjects(ctx context.Context, goldenPath, version string, staleOnly bool) ([]ProjectRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.name,
		       COALESCE(p.group_id, ''),
		       COALESCE(p.artifact_id, p.name),
		       COALESCE(p.git_repo_name, ''),
		       COALESCE(p.git_repo_url, ''),
		       COALESCE(p.gp_pin, ''),
		       COALESCE(p.version_pin, ''),
		       p.canary_mode::text,
		       COALESCE(lb.branch, ''),
		       lb.reported_at
		FROM projects p
		LEFT JOIN LATERAL (
			SELECT branch, reported_at
			FROM build_reports br
			WHERE br.project_id = p.id
			ORDER BY br.reported_at DESC
			LIMIT 1
		) lb ON true
		WHERE ($1 = '' OR p.gp_pin = $1)
		  AND ($2 = '' OR p.version_pin = $2)
		  AND (
		    NOT $3::bool OR NOT EXISTS (
		      SELECT 1 FROM build_reports br2
		      WHERE br2.project_id = p.id
		        AND br2.reported_at > now() - interval '90 days'
		    )
		  )
		ORDER BY p.name
	`, goldenPath, version, staleOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ProjectRow
	for rows.Next() {
		var row ProjectRow
		var lastBuild *time.Time
		if err := rows.Scan(
			&row.Name, &row.GroupID, &row.ArtifactID, &row.GitRepoName, &row.GitRepoURL,
			&row.GoldenPath, &row.Version, &row.CanaryMode, &row.Branch, &lastBuild,
		); err != nil {
			return nil, err
		}
		row.LastBuildAt = lastBuild
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
