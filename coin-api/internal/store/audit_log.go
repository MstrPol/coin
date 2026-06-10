package store

import (
	"context"
	"encoding/json"
	"time"
)

type AuditLogEntry struct {
	ID         int64           `json:"id"`
	Action     string          `json:"action"`
	EntityType string          `json:"entityType"`
	EntityKey  string          `json:"entityKey"`
	Actor      *string         `json:"actor,omitempty"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"createdAt"`
}

type AuditLogFilter struct {
	EntityType string
	Action     string
	Limit      int
	Offset     int
}

func (s *Store) ListAuditLog(ctx context.Context, f AuditLogFilter) ([]AuditLogEntry, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, action, entity_type, entity_key, actor, payload, created_at
		FROM audit_log
		WHERE ($1 = '' OR entity_type = $1)
		  AND ($2 = '' OR action = $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`, f.EntityType, f.Action, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AuditLogEntry
	for rows.Next() {
		var row AuditLogEntry
		if err := rows.Scan(
			&row.ID, &row.Action, &row.EntityType, &row.EntityKey,
			&row.Actor, &row.Payload, &row.CreatedAt,
		); err != nil {
			return nil, err
		}
		if len(row.Payload) == 0 {
			row.Payload = json.RawMessage(`{}`)
		}
		out = append(out, row)
	}
	if out == nil {
		out = []AuditLogEntry{}
	}
	return out, rows.Err()
}
