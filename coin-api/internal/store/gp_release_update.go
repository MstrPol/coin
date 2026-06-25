package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Store) UpdateGPReleaseDraft(ctx context.Context, in PublishGPReleaseInput) (GPReleaseRow, error) {
	if in.Name == "" || in.Version == "" {
		return GPReleaseRow{}, fmt.Errorf("name and version are required")
	}

	prep, err := s.prepareGPRelease(ctx, in)
	if err != nil {
		return GPReleaseRow{}, err
	}

	rules, err := s.loadCompatibilityRules(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	if err := validateGPReleaseComposition(s, ctx, prep, rules, func(string) ComponentResolveMode {
		return ComponentResolveAdmin
	}); err != nil {
		return GPReleaseRow{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	defer tx.Rollback(ctx)

	var releaseID int64
	var status string
	err = tx.QueryRow(ctx, `
		SELECT id, status::text FROM gp_releases
		WHERE name = $1 AND version = $2
		FOR UPDATE
	`, in.Name, in.Version).Scan(&releaseID, &status)
	if err == pgx.ErrNoRows {
		return GPReleaseRow{}, ErrNotFound
	}
	if err != nil {
		return GPReleaseRow{}, fmt.Errorf("gp release lookup: %w", err)
	}
	if status != "draft" {
		return GPReleaseRow{}, ErrGPReleaseNotDraft
	}

	_, err = tx.Exec(ctx, `DELETE FROM gp_composition WHERE gp_release_id = $1`, releaseID)
	if err != nil {
		return GPReleaseRow{}, fmt.Errorf("delete composition: %w", err)
	}

	if err := insertGPComposition(ctx, tx, releaseID, prep.storeSlots, in.Composition); err != nil {
		return GPReleaseRow{}, err
	}

	entityKey := fmt.Sprintf("%s@%s", in.Name, in.Version)
	auditPayload, _ := json.Marshal(map[string]any{
		"name":        in.Name,
		"version":     in.Version,
		"composition": in.Composition,
		"status":      "draft",
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('update_gp_draft', 'gp_release', $1, $2, $3)
	`, entityKey, nullIfEmpty(in.Actor), auditPayload)
	if err != nil {
		return GPReleaseRow{}, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return GPReleaseRow{}, err
	}

	return GPReleaseRow{Name: in.Name, Version: in.Version, Status: "draft"}, nil
}
