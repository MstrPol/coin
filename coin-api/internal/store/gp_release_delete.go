package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var ErrGPReleaseNotDraft = errors.New("only draft releases can be deleted")

func (s *Store) DeleteGPReleaseDraft(ctx context.Context, name, version, actor string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var releaseID int64
	var status string
	err = tx.QueryRow(ctx, `
		SELECT id, status::text FROM gp_releases
		WHERE name = $1 AND version = $2
		FOR UPDATE
	`, name, version).Scan(&releaseID, &status)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("gp release lookup: %w", err)
	}
	if status != "draft" {
		return ErrGPReleaseNotDraft
	}

	_, err = tx.Exec(ctx, `DELETE FROM gp_releases WHERE id = $1`, releaseID)
	if err != nil {
		return fmt.Errorf("delete gp draft: %w", err)
	}

	entityKey := fmt.Sprintf("%s@%s", name, version)
	payload, _ := json.Marshal(map[string]any{
		"name":    name,
		"version": version,
		"status":  "draft",
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('delete_gp_draft', 'gp_release', $1, $2, $3)
	`, entityKey, nullIfEmpty(actor), payload)
	if err != nil {
		return fmt.Errorf("audit log: %w", err)
	}

	return tx.Commit(ctx)
}
