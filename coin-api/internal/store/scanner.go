package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProjectScanInput struct {
	Project    string
	GoldenPath string
	Version    string
	GitURL     string
}

func (s *Store) UpsertProjectScan(ctx context.Context, in ProjectScanInput) error {
	if in.Project == "" || in.GoldenPath == "" || in.Version == "" {
		return fmt.Errorf("project, goldenPath and version are required")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
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
		return fmt.Errorf("upsert project: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO project_bindings (project_id, gp_name, gp_version, git_url, last_seen_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (project_id, gp_name, gp_version)
		DO UPDATE SET git_url = COALESCE(EXCLUDED.git_url, project_bindings.git_url),
		              last_seen_at = now()
	`, projectID, in.GoldenPath, in.Version, nullString(in.GitURL))
	if err != nil {
		return fmt.Errorf("upsert binding: %w", err)
	}

	return tx.Commit(ctx)
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (s *Store) ScannerLastSHA(ctx context.Context, repo string) (string, bool, error) {
	var sha string
	err := s.pool.QueryRow(ctx, `SELECT last_sha FROM scanner_state WHERE repo_full_name = $1`, repo).Scan(&sha)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return sha, true, nil
}

func (s *Store) SaveScannerSHA(ctx context.Context, repo, sha string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO scanner_state (repo_full_name, last_sha, last_scan_at)
		VALUES ($1, $2, now())
		ON CONFLICT (repo_full_name) DO UPDATE
		SET last_sha = EXCLUDED.last_sha, last_scan_at = now()
	`, repo, sha)
	return err
}

type ScanResult struct {
	ReposTotal   int       `json:"reposTotal"`
	ReposScanned int       `json:"reposScanned"`
	ReposSkipped int       `json:"reposSkipped"`
	ReposFailed  int       `json:"reposFailed"`
	StartedAt    time.Time `json:"startedAt"`
	FinishedAt   time.Time `json:"finishedAt"`
}
