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
	return s.insertComponentVersion(ctx, in, "published", "publish_component_version")
}

func (s *Store) UpdateComponentVersionRefs(ctx context.Context, typ, name, version string, metadata, contentRef json.RawMessage) error {
	if typ == "" || name == "" || version == "" {
		return fmt.Errorf("type, name and version are required")
	}
	if err := s.requireComponentVersionDraft(ctx, typ, name, version); err != nil {
		return err
	}
	if err := validateContentRefOnWrite(contentRef); err != nil {
		return err
	}
	if len(metadata) == 0 && len(contentRef) == 0 {
		return fmt.Errorf("metadata or contentRef is required")
	}
	if len(metadata) > 0 && len(contentRef) > 0 {
		tag, err := s.pool.Exec(ctx, `
			UPDATE component_versions cv
			SET metadata = $4, content_ref = $5
			FROM components c
			WHERE cv.component_id = c.id
			  AND c.type = $1 AND c.name = $2 AND cv.version = $3
		`, typ, name, version, metadata, nullableJSON(contentRef))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("component version not found: %s/%s@%s", typ, name, version)
		}
		return nil
	}
	if len(contentRef) > 0 {
		return s.UpdateComponentVersionContentRef(ctx, typ, name, version, contentRef)
	}
	meta := metadata
	if len(meta) == 0 {
		meta = json.RawMessage(`{}`)
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE component_versions cv
		SET metadata = $4
		FROM components c
		WHERE cv.component_id = c.id
		  AND c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version, meta)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("component version not found: %s/%s@%s", typ, name, version)
	}
	return nil
}

func (s *Store) UpdateComponentVersionContentRef(ctx context.Context, typ, name, version string, contentRef json.RawMessage) error {
	if typ == "" || name == "" || version == "" {
		return fmt.Errorf("type, name and version are required")
	}
	if err := s.requireComponentVersionDraft(ctx, typ, name, version); err != nil {
		return err
	}
	if err := validateContentRefOnWrite(contentRef); err != nil {
		return err
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE component_versions cv
		SET content_ref = $4
		FROM components c
		WHERE cv.component_id = c.id
		  AND c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version, nullableJSON(contentRef))
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
