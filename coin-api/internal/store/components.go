package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrDuplicateVersion  = errors.New("component version already exists")
	ErrDuplicateComponent = errors.New("component already exists")
)

type ComponentVersionInput struct {
	Type       string
	Name       string
	Version    string
	Metadata   json.RawMessage
	ContentRef json.RawMessage
	Actor      string
}

type ComponentVersionRow struct {
	ID           int64
	ComponentID  int64
	Type         string
	Name         string
	Version      string
	Status       string
	Metadata     json.RawMessage
	ContentRef   json.RawMessage
}

func (s *Store) CreateComponent(ctx context.Context, typ, name, actor string) error {
	if typ == "" || name == "" {
		return fmt.Errorf("type and name are required")
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO components (type, name) VALUES ($1, $2)`, typ, name)
	if isUniqueViolation(err) {
		return ErrDuplicateComponent
	}
	if err != nil {
		return err
	}
	entityKey := fmt.Sprintf("%s/%s", typ, name)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('create_component', 'component', $1, $2, '{}'::jsonb)
	`, entityKey, nullIfEmpty(actor))
	return err
}

func (s *Store) PublishComponentVersion(ctx context.Context, in ComponentVersionInput) (ComponentVersionRow, error) {
	if in.Type == "" || in.Name == "" || in.Version == "" {
		return ComponentVersionRow{}, fmt.Errorf("type, name and version are required")
	}
	meta := in.Metadata
	if len(meta) == 0 {
		meta = json.RawMessage(`{}`)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ComponentVersionRow{}, err
	}
	defer tx.Rollback(ctx)

	var componentID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO components (type, name) VALUES ($1, $2)
		ON CONFLICT (type, name) DO UPDATE SET type=EXCLUDED.type
		RETURNING id
	`, in.Type, in.Name).Scan(&componentID)
	if err != nil {
		return ComponentVersionRow{}, fmt.Errorf("component upsert: %w", err)
	}

	var versionID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO component_versions (component_id, version, metadata, content_ref)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, componentID, in.Version, meta, nullableJSON(in.ContentRef)).Scan(&versionID)
	if err != nil {
		if isUniqueViolation(err) {
			return ComponentVersionRow{}, ErrDuplicateVersion
		}
		return ComponentVersionRow{}, fmt.Errorf("component version insert: %w", err)
	}

	entityKey := fmt.Sprintf("%s/%s@%s", in.Type, in.Name, in.Version)
	auditPayload, _ := json.Marshal(map[string]any{
		"type":       in.Type,
		"name":       in.Name,
		"version":    in.Version,
		"metadata":   json.RawMessage(meta),
		"contentRef": json.RawMessage(in.ContentRef),
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('publish_component_version', 'component_version', $1, $2, $3)
	`, entityKey, nullIfEmpty(in.Actor), auditPayload)
	if err != nil {
		return ComponentVersionRow{}, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return ComponentVersionRow{}, err
	}

	return ComponentVersionRow{
		ID:          versionID,
		ComponentID: componentID,
		Type:        in.Type,
		Name:        in.Name,
		Version:     in.Version,
		Status:      "published",
		Metadata:    meta,
		ContentRef:  in.ContentRef,
	}, nil
}

func (s *Store) UpdateComponentVersionRefs(ctx context.Context, typ, name, version string, metadata, contentRef json.RawMessage) error {
	if typ == "" || name == "" || version == "" {
		return fmt.Errorf("type, name and version are required")
	}
	meta := metadata
	if len(meta) == 0 {
		meta = json.RawMessage(`{}`)
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE component_versions cv
		SET metadata = $4, content_ref = $5
		FROM components c
		WHERE cv.component_id = c.id
		  AND c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version, meta, nullableJSON(contentRef))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("component version not found: %s/%s@%s", typ, name, version)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func nullableJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
