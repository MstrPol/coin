package store

import (
	"context"
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
	Project    string
	GoldenPath string
	Result     string
	Limit      int
	Offset     int
}

func (s *Store) ListBuildReports(ctx context.Context, f ListBuildReportsFilter) ([]BuildReportRow, error) {
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT br.id, p.name, br.gp_name, br.gp_version,
		       COALESCE(br.resolved_version, br.gp_version),
		       COALESCE(br.branch, ''), COALESCE(br.build_url, ''),
		       br.result, COALESCE(br.channel, ''), COALESCE(br.failed_stage, ''),
		       br.reported_at
		FROM build_reports br
		JOIN projects p ON p.id = br.project_id
		WHERE 1=1
	`
	args := []any{}
	n := 1
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
	query += fmt.Sprintf(" ORDER BY br.reported_at DESC LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	return items, rows.Err()
}
