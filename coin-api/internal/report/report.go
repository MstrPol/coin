package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Input struct {
	Project         string
	GoldenPath      string
	Version         string
	Branch          string
	BuildURL        string
	Result          string
	ManifestHash    string
	GitURL          string
	Channel         string
	RequestedPin    string
	FailedStage     string
	ResolvedVersion string
}

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func NormalizeResult(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "SUCCESS":
		return "success"
	case "FAILURE":
		return "failure"
	case "ABORTED", "UNSTABLE":
		return "aborted"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func (s *Service) Save(ctx context.Context, in Input) (int64, error) {
	if in.Project == "" || in.GoldenPath == "" || in.Version == "" || in.Result == "" {
		return 0, fmt.Errorf("project, goldenPath, version, result are required")
	}
	in.Result = NormalizeResult(in.Result)
	if in.Result == "" {
		return 0, fmt.Errorf("invalid result")
	}
	if in.Channel == "" {
		in.Channel = "stable"
	}
	if in.ResolvedVersion == "" {
		in.ResolvedVersion = in.Version
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var projectID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO projects (name, updated_at)
		VALUES ($1, now())
		ON CONFLICT (name) DO UPDATE SET updated_at = now()
		RETURNING id
	`, in.Project).Scan(&projectID)
	if err != nil {
		return 0, fmt.Errorf("upsert project: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO project_bindings (project_id, gp_name, gp_version, git_url, last_seen_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (project_id, gp_name, gp_version)
		DO UPDATE SET git_url = COALESCE(EXCLUDED.git_url, project_bindings.git_url),
		              last_seen_at = now()
	`, projectID, in.GoldenPath, in.Version, nullIfEmpty(in.GitURL))
	if err != nil {
		return 0, fmt.Errorf("upsert binding: %w", err)
	}

	var reportID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO build_reports (
			project_id, gp_name, gp_version, branch, build_url, result, manifest_hash,
			channel, requested_pin, failed_stage, resolved_version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, projectID, in.GoldenPath, in.Version,
		nullIfEmpty(in.Branch), nullIfEmpty(in.BuildURL), in.Result, nullIfEmpty(in.ManifestHash),
		in.Channel, nullIfEmpty(in.RequestedPin), nullIfEmpty(in.FailedStage), in.ResolvedVersion,
	).Scan(&reportID)
	if err != nil {
		return 0, fmt.Errorf("insert report: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return reportID, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func (s *Service) LastReportAge(ctx context.Context, project string) (time.Duration, error) {
	var reportedAt time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT br.reported_at
		FROM build_reports br
		JOIN projects p ON p.id = br.project_id
		WHERE p.name = $1
		ORDER BY br.reported_at DESC
		LIMIT 1
	`, project).Scan(&reportedAt)
	if err != nil {
		return 0, err
	}
	return time.Since(reportedAt), nil
}
