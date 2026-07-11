package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func componentVersionEntityKey(typ, name, version string) string {
	return fmt.Sprintf("%s/%s@%s", typ, name, version)
}

func (s *Store) DeleteComponentVersionDraft(ctx context.Context, typ, name, version, actor string) error {
	if typ == "" || name == "" || version == "" {
		return fmt.Errorf("type, name and version are required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var versionID int64
	var status string
	err = tx.QueryRow(ctx, `
		SELECT cv.id, cv.status::text
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
		FOR UPDATE
	`, typ, name, version).Scan(&versionID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("component version lookup: %w", err)
	}
	if status != "draft" {
		return ErrComponentVersionNotDraft
	}

	_, err = tx.Exec(ctx, `DELETE FROM component_versions WHERE id = $1`, versionID)
	if err != nil {
		return fmt.Errorf("delete component version: %w", err)
	}

	entityKey := componentVersionEntityKey(typ, name, version)
	payload, _ := json.Marshal(map[string]any{
		"type":    typ,
		"name":    name,
		"version": version,
		"status":  "draft",
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('delete_component_draft', 'component_version', $1, $2, $3)
	`, entityKey, nullIfEmpty(actor), payload)
	if err != nil {
		return fmt.Errorf("audit log: %w", err)
	}

	return tx.Commit(ctx)
}
