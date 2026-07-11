package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/canary"
)

type CanaryPolicyRow struct {
	GPName                     string `json:"gpName"`
	Enabled                    bool   `json:"enabled"`
	CanaryPercent              int    `json:"canaryPercent"`
	DegradedThresholdPct       int    `json:"degradedThresholdPct"`
	CriticalThresholdPct       int    `json:"criticalThresholdPct"`
	CriticalConsecutiveFailures int `json:"criticalConsecutiveFailures"`
}

func (s *Store) GetCanaryPolicy(ctx context.Context, gpName string) (CanaryPolicyRow, error) {
	var row CanaryPolicyRow
	err := s.pool.QueryRow(ctx, `
		SELECT gp_name, enabled, canary_percent, degraded_threshold_pct,
		       critical_threshold_pct, critical_consecutive_failures
		FROM canary_policy WHERE gp_name = $1
	`, gpName).Scan(
		&row.GPName, &row.Enabled, &row.CanaryPercent,
		&row.DegradedThresholdPct, &row.CriticalThresholdPct, &row.CriticalConsecutiveFailures,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return CanaryPolicyRow{GPName: gpName}, nil
	}
	return row, err
}

func (s *Store) UpsertCanaryPolicy(ctx context.Context, row CanaryPolicyRow, actor string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO canary_policy (
			gp_name, enabled, canary_percent,
			degraded_threshold_pct, critical_threshold_pct, critical_consecutive_failures
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (gp_name) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			canary_percent = EXCLUDED.canary_percent,
			degraded_threshold_pct = EXCLUDED.degraded_threshold_pct,
			critical_threshold_pct = EXCLUDED.critical_threshold_pct,
			critical_consecutive_failures = EXCLUDED.critical_consecutive_failures
	`, row.GPName, row.Enabled, row.CanaryPercent,
		row.DegradedThresholdPct, row.CriticalThresholdPct, row.CriticalConsecutiveFailures)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('update_canary_policy', 'canary_policy', $1, $2, '{}'::jsonb)
	`, row.GPName, nullIfEmpty(actor))
	return err
}

func (s *Store) GetProjectCanaryMode(ctx context.Context, projectName string) (string, error) {
	if projectName == "" {
		return "default", nil
	}
	var mode string
	err := s.pool.QueryRow(ctx, `SELECT canary_mode::text FROM projects WHERE name=$1`, projectName).Scan(&mode)
	if errors.Is(err, pgx.ErrNoRows) {
		return "default", nil
	}
	return mode, err
}

func (s *Store) SetProjectCanaryMode(ctx context.Context, projectName, mode, actor string) error {
	if projectName == "" || mode == "" {
		return errors.New("project and mode are required")
	}
	switch mode {
	case "default", "canary", "stable":
	default:
		return errors.New("mode must be default, canary, or stable")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO projects (name, canary_mode, updated_at)
		VALUES ($1, $2::canary_mode, now())
		ON CONFLICT (name) DO UPDATE SET canary_mode = EXCLUDED.canary_mode, updated_at = now()
	`, projectName, mode)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('update_project_canary_mode', 'project', $1, $2, jsonb_build_object('mode', $3))
	`, projectName, nullIfEmpty(actor), mode)
	return err
}

type FailureEntry struct {
	Project     string    `json:"project"`
	FailedStage *string   `json:"failedStage,omitempty"`
	BuildURL    *string   `json:"buildUrl,omitempty"`
	ReportedAt  time.Time `json:"reportedAt"`
}

type HealthSummary struct {
	GPName         string         `json:"gpName"`
	Version        string         `json:"version"`
	Channel        string         `json:"channel"`
	WindowHours    int            `json:"windowHours"`
	SuccessCount   int            `json:"successCount"`
	FailureCount   int            `json:"failureCount"`
	FailureRate    float64        `json:"failureRate"`
	Health         string         `json:"health"`
	RecentFailures []FailureEntry `json:"recentFailures"`
}

func (s *Store) AggregateHealth(ctx context.Context, gpName, version, channel string, window time.Duration, policy CanaryPolicyRow) (HealthSummary, error) {
	windowHours := int(window.Hours())
	if windowHours < 1 {
		windowHours = 24
	}

	var successCount, failureCount int
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE result = 'success')::int,
			COUNT(*) FILTER (WHERE result = 'failure')::int
		FROM build_reports
		WHERE gp_name = $1 AND gp_version = $2
		  AND COALESCE(channel, 'stable') = $3
		  AND reported_at >= now() - $4::interval
	`, gpName, version, channel, window.String()).Scan(&successCount, &failureCount)
	if err != nil {
		return HealthSummary{}, err
	}

	total := successCount + failureCount
	var failureRate float64
	if total > 0 {
		failureRate = float64(failureCount) / float64(total) * 100
	}

	health := classifyHealth(failureRate, failureCount, policy)

	rows, err := s.pool.Query(ctx, `
		SELECT p.name, br.failed_stage, br.build_url, br.reported_at
		FROM build_reports br
		JOIN projects p ON p.id = br.project_id
		WHERE br.gp_name = $1 AND br.gp_version = $2
		  AND COALESCE(br.channel, 'stable') = $3
		  AND br.result = 'failure'
		  AND br.reported_at >= now() - $4::interval
		ORDER BY br.reported_at DESC
		LIMIT 10
	`, gpName, version, channel, window.String())
	if err != nil {
		return HealthSummary{}, err
	}
	defer rows.Close()

	var failures []FailureEntry
	for rows.Next() {
		var f FailureEntry
		if err := rows.Scan(&f.Project, &f.FailedStage, &f.BuildURL, &f.ReportedAt); err != nil {
			return HealthSummary{}, err
		}
		failures = append(failures, f)
	}
	if failures == nil {
		failures = []FailureEntry{}
	}

	return HealthSummary{
		GPName:         gpName,
		Version:        version,
		Channel:        channel,
		WindowHours:    windowHours,
		SuccessCount:   successCount,
		FailureCount:   failureCount,
		FailureRate:    failureRate,
		Health:         health,
		RecentFailures: failures,
	}, rows.Err()
}

func classifyHealth(failureRate float64, failureCount int, policy CanaryPolicyRow) string {
	if failureCount == 0 {
		return "healthy"
	}
	degraded := policy.DegradedThresholdPct
	if degraded == 0 {
		degraded = 10
	}
	critical := policy.CriticalThresholdPct
	if critical == 0 {
		critical = 25
	}
	consec := policy.CriticalConsecutiveFailures
	if consec == 0 {
		consec = 3
	}
	if failureRate > float64(critical) || failureCount >= consec {
		return "critical"
	}
	if failureRate > float64(degraded) {
		return "degraded"
	}
	return "healthy"
}

// CountProjectsInCanaryBucket counts how many projects would get canary at given percent.
func (s *Store) CountProjectsInCanaryBucket(ctx context.Context, percent int) (inCanary, total int, err error) {
	rows, err := s.pool.Query(ctx, `SELECT name, canary_mode::text FROM projects ORDER BY name`)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, mode string
		if err := rows.Scan(&name, &mode); err != nil {
			return 0, 0, err
		}
		total++
		switch mode {
		case "canary":
			inCanary++
		case "stable":
			// no-op
		default:
			if canary.ProjectBucket(name) < percent {
				inCanary++
			}
		}
	}
	return inCanary, total, rows.Err()
}
