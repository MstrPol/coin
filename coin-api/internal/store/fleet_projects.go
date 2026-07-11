package store

import (
	"context"
	"encoding/csv"
	"fmt"
	"time"
)

type ListProjectsFilter struct {
	GoldenPath string
	Version    string
	StaleOnly  bool
	Limit      int
	Offset     int
}

func normalizeProjectsLimitOffset(f *ListProjectsFilter) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
}

const projectsWhereSQL = `
		WHERE ($1 = '' OR p.gp_pin = $1)
		  AND ($2 = '' OR p.version_pin = $2)
		  AND (
		    NOT $3::bool OR NOT EXISTS (
		      SELECT 1 FROM build_reports br2
		      WHERE br2.project_id = p.id
		        AND br2.reported_at > now() - interval '90 days'
		    )
		  )
`

func (s *Store) CountProjects(ctx context.Context, f ListProjectsFilter) (int, error) {
	var total int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM projects p
	`+projectsWhereSQL, f.GoldenPath, f.Version, f.StaleOnly).Scan(&total)
	return total, err
}

func (s *Store) ListProjects(ctx context.Context, f ListProjectsFilter) ([]ProjectRow, error) {
	normalizeProjectsLimitOffset(&f)

	query := `
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
	` + projectsWhereSQL + `
		ORDER BY p.name
		LIMIT $4 OFFSET $5
	`

	rows, err := s.pool.Query(ctx, query, f.GoldenPath, f.Version, f.StaleOnly, f.Limit, f.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProjectRows(rows)
}

func (s *Store) WriteProjectsCSV(ctx context.Context, f ListProjectsFilter, w *csv.Writer) error {
	query := `
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
	` + projectsWhereSQL + `
		ORDER BY p.name
	`

	rows, err := s.pool.Query(ctx, query, f.GoldenPath, f.Version, f.StaleOnly)
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := w.Write([]string{
		"name", "groupId", "artifactId", "gitRepoName", "gitRepoUrl",
		"goldenPath", "version", "canaryMode", "branch", "lastBuildAt",
	}); err != nil {
		return err
	}

	for rows.Next() {
		var row ProjectRow
		var lastBuild *time.Time
		if err := rows.Scan(
			&row.Name, &row.GroupID, &row.ArtifactID, &row.GitRepoName, &row.GitRepoURL,
			&row.GoldenPath, &row.Version, &row.CanaryMode, &row.Branch, &lastBuild,
		); err != nil {
			return err
		}
		lastBuildStr := ""
		if lastBuild != nil {
			lastBuildStr = lastBuild.UTC().Format(time.RFC3339)
		}
		if err := w.Write([]string{
			row.Name, row.GroupID, row.ArtifactID, row.GitRepoName, row.GitRepoURL,
			row.GoldenPath, row.Version, row.CanaryMode, row.Branch, lastBuildStr,
		}); err != nil {
			return err
		}
	}
	return rows.Err()
}

type projectRowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanProjectRows(rows projectRowScanner) ([]ProjectRow, error) {
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
	if out == nil {
		out = []ProjectRow{}
	}
	return out, rows.Err()
}

func FormatProjectExportFilename() string {
	return fmt.Sprintf("projects-%s.csv", time.Now().UTC().Format("20060102"))
}
