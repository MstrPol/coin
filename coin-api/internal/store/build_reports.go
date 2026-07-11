package store

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"
)

type BuildReportRow struct {
	ID              int64     `json:"id"`
	Project         string    `json:"project"`
	GoldenPath      string    `json:"goldenPath"`
	Version         string    `json:"version"`
	ResolvedVersion string    `json:"resolvedVersion,omitempty"`
	Branch          string    `json:"branch,omitempty"`
	BuildURL        string    `json:"buildUrl,omitempty"`
	Result          string    `json:"result"`
	Channel         string    `json:"channel,omitempty"`
	FailedStage     string    `json:"failedStage,omitempty"`
	ReportedAt      time.Time `json:"reportedAt"`
}

type ListBuildReportsFilter struct {
	Project        string
	GoldenPath     string
	Result         string
	ReportedAfter  *time.Time
	ReportedBefore *time.Time
	Limit          int
	Offset         int
}

func normalizeBuildReportsLimitOffset(f *ListBuildReportsFilter) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
}

func appendBuildReportsWhere(query string, f ListBuildReportsFilter, args []any) (string, []any) {
	n := len(args) + 1
	if f.Project != "" {
		query += fmt.Sprintf(" AND p.name = $%d", n)
		args = append(args, f.Project)
		n++
	}
	if f.GoldenPath != "" {
		query += fmt.Sprintf(" AND br.gp_name = $%d", n)
		args = append(args, f.GoldenPath)
		n++
	}
	if f.Result != "" {
		query += fmt.Sprintf(" AND br.result = $%d", n)
		args = append(args, strings.ToLower(f.Result))
		n++
	}
	if f.ReportedAfter != nil {
		query += fmt.Sprintf(" AND br.reported_at >= $%d", n)
		args = append(args, *f.ReportedAfter)
		n++
	}
	if f.ReportedBefore != nil {
		query += fmt.Sprintf(" AND br.reported_at <= $%d", n)
		args = append(args, *f.ReportedBefore)
		n++
	}
	return query, args
}

const buildReportsFromSQL = `
		FROM build_reports br
		JOIN projects p ON p.id = br.project_id
		WHERE 1=1
`

func (s *Store) CountBuildReports(ctx context.Context, f ListBuildReportsFilter) (int, error) {
	query := `SELECT COUNT(*)::int ` + buildReportsFromSQL
	query, args := appendBuildReportsWhere(query, f, nil)

	var total int
	err := s.pool.QueryRow(ctx, query, args...).Scan(&total)
	return total, err
}

func (s *Store) ListBuildReports(ctx context.Context, f ListBuildReportsFilter) ([]BuildReportRow, error) {
	normalizeBuildReportsLimitOffset(&f)

	query := `
		SELECT br.id, p.name, br.gp_name, br.gp_version,
		       COALESCE(br.resolved_version, br.gp_version),
		       COALESCE(br.branch, ''), COALESCE(br.build_url, ''),
		       br.result, COALESCE(br.channel, ''), COALESCE(br.failed_stage, ''),
		       br.reported_at
	` + buildReportsFromSQL

	args := []any{}
	query, args = appendBuildReportsWhere(query, f, args)
	n := len(args) + 1
	query += fmt.Sprintf(" ORDER BY br.reported_at DESC LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBuildReportRows(rows)
}

func (s *Store) WriteBuildReportsCSV(ctx context.Context, f ListBuildReportsFilter, w *csv.Writer) error {
	query := `
		SELECT br.id, p.name, br.gp_name, br.gp_version,
		       COALESCE(br.resolved_version, br.gp_version),
		       COALESCE(br.branch, ''), COALESCE(br.build_url, ''),
		       br.result, COALESCE(br.channel, ''), COALESCE(br.failed_stage, ''),
		       br.reported_at
	` + buildReportsFromSQL

	query, args := appendBuildReportsWhere(query, f, nil)
	query += " ORDER BY br.reported_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := w.Write([]string{
		"id", "project", "goldenPath", "version", "resolvedVersion",
		"result", "channel", "branch", "buildUrl", "failedStage", "reportedAt",
	}); err != nil {
		return err
	}

	for rows.Next() {
		var row BuildReportRow
		if err := rows.Scan(
			&row.ID, &row.Project, &row.GoldenPath, &row.Version,
			&row.ResolvedVersion, &row.Branch, &row.BuildURL,
			&row.Result, &row.Channel, &row.FailedStage, &row.ReportedAt,
		); err != nil {
			return err
		}
		if err := w.Write([]string{
			fmt.Sprintf("%d", row.ID),
			row.Project,
			row.GoldenPath,
			row.Version,
			row.ResolvedVersion,
			row.Result,
			row.Channel,
			row.Branch,
			row.BuildURL,
			row.FailedStage,
			row.ReportedAt.UTC().Format(time.RFC3339),
		}); err != nil {
			return err
		}
	}
	return rows.Err()
}

func scanBuildReportRows(rows projectRowScanner) ([]BuildReportRow, error) {
	var items []BuildReportRow
	for rows.Next() {
		var row BuildReportRow
		if err := rows.Scan(
			&row.ID, &row.Project, &row.GoldenPath, &row.Version,
			&row.ResolvedVersion, &row.Branch, &row.BuildURL,
			&row.Result, &row.Channel, &row.FailedStage, &row.ReportedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	if items == nil {
		items = []BuildReportRow{}
	}
	return items, rows.Err()
}

func FormatBuildReportsExportFilename() string {
	return fmt.Sprintf("build-reports-%s.csv", time.Now().UTC().Format("20060102"))
}
