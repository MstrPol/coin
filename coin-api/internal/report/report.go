package report

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Input struct {
	Project         string
	GroupID         string
	ArtifactID      string
	GoldenPath      string
	Version         string
	ConfigVersion   string
	Branch          string
	BuildURL        string
	Result          string
	ManifestHash    string
	GitURL          string
	Channel         string
	RequestedPin    string
	FailedStage     string
	ResolvedVersion string
	Outputs         []map[string]any
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

func gitRepoName(gitURL string) string {
	gitURL = strings.TrimSpace(gitURL)
	if gitURL == "" {
		return ""
	}
	gitURL = strings.TrimSuffix(gitURL, ".git")
	return path.Base(gitURL)
}

func (s *Service) Save(ctx context.Context, in Input) (int64, error) {
	if in.Project == "" || in.GoldenPath == "" || in.Result == "" {
		return 0, fmt.Errorf("project, goldenPath, result are required")
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
	if in.ArtifactID == "" {
		in.ArtifactID = in.Project
	}
	gpVersion := in.Version
	if gpVersion == "" {
		gpVersion = in.ResolvedVersion
	}
	configPin := in.ConfigVersion
	if configPin == "" {
		configPin = in.RequestedPin
	}
	repoName := gitRepoName(in.GitURL)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var projectID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO projects (
			name, group_id, artifact_id, git_repo_name, git_repo_url,
			gp_pin, version_pin, last_build_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
		ON CONFLICT (name) DO UPDATE SET
			group_id = COALESCE(NULLIF(EXCLUDED.group_id, ''), projects.group_id),
			artifact_id = COALESCE(NULLIF(EXCLUDED.artifact_id, ''), projects.artifact_id),
			git_repo_name = COALESCE(NULLIF(EXCLUDED.git_repo_name, ''), projects.git_repo_name),
			git_repo_url = COALESCE(NULLIF(EXCLUDED.git_repo_url, ''), projects.git_repo_url),
			gp_pin = COALESCE(NULLIF(EXCLUDED.gp_pin, ''), projects.gp_pin),
			version_pin = COALESCE(NULLIF(EXCLUDED.version_pin, ''), projects.version_pin),
			last_build_at = now(),
			updated_at = now()
		RETURNING id
	`, in.Project, nullIfEmpty(in.GroupID), in.ArtifactID,
		nullIfEmpty(repoName), nullIfEmpty(in.GitURL),
		in.GoldenPath, nullIfEmpty(configPin)).Scan(&projectID)
	if err != nil {
		return 0, fmt.Errorf("upsert project: %w", err)
	}

	outputsRaw, err := json.Marshal(in.Outputs)
	if err != nil {
		return 0, fmt.Errorf("marshal outputs: %w", err)
	}
	if len(in.Outputs) == 0 {
		outputsRaw = []byte("[]")
	}

	var reportID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO build_reports (
			project_id, gp_name, gp_version, branch, build_url, result, manifest_hash,
			channel, requested_pin, failed_stage, resolved_version, outputs
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb)
		RETURNING id
	`, projectID, in.GoldenPath, gpVersion,
		nullIfEmpty(in.Branch), nullIfEmpty(in.BuildURL), in.Result, nullIfEmpty(in.ManifestHash),
		in.Channel, nullIfEmpty(configPin), nullIfEmpty(in.FailedStage), in.ResolvedVersion, outputsRaw,
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
