package store

import (
	"context"
	"time"

	"coin.local/coin-api/internal/manifest"
)

type DashboardStats struct {
	Projects      int `json:"projects"`
	StaleProjects int `json:"staleProjects"`
	GPReleases    int `json:"gpReleases"`
	BuildReports  int `json:"buildReports"`
	GoldenPaths   int `json:"goldenPaths"`
}

type ProjectRow struct {
	Name        string     `json:"name"`
	GroupID     string     `json:"groupId,omitempty"`
	ArtifactID  string     `json:"artifactId,omitempty"`
	GitRepoName string     `json:"gitRepoName,omitempty"`
	GitRepoURL  string     `json:"gitRepoUrl,omitempty"`
	GoldenPath  string     `json:"goldenPath"`
	Version     string     `json:"version"`
	CanaryMode  string     `json:"canaryMode"`
	Branch      string     `json:"branch,omitempty"`
	LastBuildAt *time.Time `json:"lastBuildAt,omitempty"`
}

type GPReleaseListItem struct {
	Name         string                `json:"name"`
	Version      string                `json:"version"`
	Status       string                `json:"status"`
	Destinations manifest.Destinations `json:"destinations"`
	ManifestHash *string               `json:"manifestHash,omitempty"`
	ManifestURL  *string               `json:"manifestUrl,omitempty"`
	GitExportTag *string               `json:"gitExportTag,omitempty"`
	CreatedAt    time.Time             `json:"createdAt"`
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

func (s *Store) ListGPReleases(ctx context.Context, name string, includeDrafts bool) ([]GPReleaseListItem, error) {
	query := `
		SELECT name, version, status, image_registry_prefix, build_cache_enabled, artifact_repository_base, manifest_hash, manifest_url, git_export_tag, created_at
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
			&row.Destinations.ImageRegistryPrefix,
			&row.Destinations.BuildCacheEnabled,
			&row.Destinations.ArtifactRepositoryBase,
			&row.ManifestHash, &row.ManifestURL, &row.GitExportTag, &row.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
