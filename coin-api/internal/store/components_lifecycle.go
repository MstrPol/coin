package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/componentpackage"
)

var (
	ErrInvalidComponentStatusTransition = errors.New("invalid component status transition")
	ErrComponentVersionNotDraft         = errors.New("component version is not draft")
	ErrComponentVersionNotCanary          = errors.New("component version is not canary")
	ErrComponentPackageNotRegistered      = errors.New("component package is not registered (content_ref v2 required)")
)

func (s *Store) CreateDraftComponentVersion(ctx context.Context, in ComponentVersionInput) (ComponentVersionRow, error) {
	return s.insertComponentVersion(ctx, in, "draft", "create_component_draft")
}

func (s *Store) PublishComponentToCanary(ctx context.Context, typ, name, version, actor string) (ComponentVersionRow, error) {
	return s.transitionComponentVersion(ctx, typ, name, version, actor, "draft", "canary", "publish_component_canary")
}

func (s *Store) PromoteComponentToPublished(ctx context.Context, typ, name, version, actor string) (ComponentVersionRow, error) {
	return s.transitionComponentVersion(ctx, typ, name, version, actor, "draft", "published", "publish_component_version")
}

func (s *Store) insertComponentVersion(ctx context.Context, in ComponentVersionInput, status, auditAction string) (ComponentVersionRow, error) {
	if in.Type == "" || in.Name == "" || in.Version == "" {
		return ComponentVersionRow{}, fmt.Errorf("type, name and version are required")
	}
	if err := validateContentRefOnWriteForType(in.Type, in.ContentRef); err != nil {
		return ComponentVersionRow{}, err
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
		INSERT INTO component_versions (component_id, version, status, metadata, content_ref)
		VALUES ($1, $2, $3::component_status, $4, $5)
		RETURNING id
	`, componentID, in.Version, status, meta, nullableJSON(in.ContentRef)).Scan(&versionID)
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
		"status":     status,
		"metadata":   json.RawMessage(meta),
		"contentRef": json.RawMessage(in.ContentRef),
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ($1, 'component_version', $2, $3, $4)
	`, auditAction, entityKey, nullIfEmpty(in.Actor), auditPayload)
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
		Status:      status,
		Metadata:    meta,
		ContentRef:  in.ContentRef,
	}, nil
}

func (s *Store) PromoteComponentToPublishedWithContentRef(ctx context.Context, typ, name, version string, contentRef json.RawMessage, actor string) (ComponentVersionRow, error) {
	if typ == "" || name == "" || version == "" {
		return ComponentVersionRow{}, fmt.Errorf("type, name and version are required")
	}
	if _, err := componentpackage.ValidateContentRefV2(contentRef); err != nil {
		return ComponentVersionRow{}, fmt.Errorf("%w: %s", ErrInvalidContentRef, err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ComponentVersionRow{}, err
	}
	defer tx.Rollback(ctx)

	var row ComponentVersionRow
	var meta, existingRef []byte
	err = tx.QueryRow(ctx, `
		SELECT cv.id, c.id, c.type, c.name, cv.version, cv.status::text, cv.metadata, cv.content_ref
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
		FOR UPDATE
	`, typ, name, version).Scan(
		&row.ID, &row.ComponentID, &row.Type, &row.Name, &row.Version, &row.Status, &meta, &existingRef,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ComponentVersionRow{}, ErrNotFound
		}
		return ComponentVersionRow{}, err
	}
	if row.Status != "draft" && row.Status != "canary" {
		return ComponentVersionRow{}, ErrComponentVersionNotDraft
	}

	_, err = tx.Exec(ctx, `
		UPDATE component_versions SET status = 'published'::component_status, content_ref = $2
		WHERE id = $1
	`, row.ID, nullableJSON(contentRef))
	if err != nil {
		return ComponentVersionRow{}, err
	}

	entityKey := fmt.Sprintf("%s/%s@%s", typ, name, version)
	auditPayload, _ := json.Marshal(map[string]any{
		"type":    typ,
		"name":    name,
		"version": version,
		"from":    row.Status,
		"to":      "published",
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('promote_component_version', 'component_version', $1, $2, $3)
	`, entityKey, nullIfEmpty(actor), auditPayload)
	if err != nil {
		return ComponentVersionRow{}, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return ComponentVersionRow{}, err
	}

	row.Status = "published"
	row.Metadata = meta
	row.ContentRef = contentRef
	return row, nil
}

func (s *Store) transitionComponentVersion(ctx context.Context, typ, name, version, actor, fromStatus, toStatus, auditAction string) (ComponentVersionRow, error) {
	if typ == "" || name == "" || version == "" {
		return ComponentVersionRow{}, fmt.Errorf("type, name and version are required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ComponentVersionRow{}, err
	}
	defer tx.Rollback(ctx)

	var row ComponentVersionRow
	var meta, contentRef []byte
	err = tx.QueryRow(ctx, `
		SELECT cv.id, c.id, c.type, c.name, cv.version, cv.status::text, cv.metadata, cv.content_ref
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
		FOR UPDATE
	`, typ, name, version).Scan(
		&row.ID, &row.ComponentID, &row.Type, &row.Name, &row.Version, &row.Status, &meta, &contentRef,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ComponentVersionRow{}, ErrNotFound
		}
		return ComponentVersionRow{}, err
	}
	if row.Status != fromStatus {
		if fromStatus == "draft" && row.Status != "draft" {
			return ComponentVersionRow{}, ErrComponentVersionNotDraft
		}
		if fromStatus == "canary" && row.Status != "canary" {
			return ComponentVersionRow{}, ErrComponentVersionNotCanary
		}
		return ComponentVersionRow{}, fmt.Errorf("%w: %s -> %s", ErrInvalidComponentStatusTransition, row.Status, toStatus)
	}
	if toStatus == "canary" && !componentpackage.IsRegisteredForCanary(contentRef) {
		return ComponentVersionRow{}, ErrComponentPackageNotRegistered
	}

	_, err = tx.Exec(ctx, `
		UPDATE component_versions SET status = $4::component_status
		WHERE id = $1 AND component_id = $2 AND version = $3
	`, row.ID, row.ComponentID, row.Version, toStatus)
	if err != nil {
		return ComponentVersionRow{}, err
	}

	entityKey := fmt.Sprintf("%s/%s@%s", typ, name, version)
	auditPayload, _ := json.Marshal(map[string]any{
		"type":    typ,
		"name":    name,
		"version": version,
		"from":    fromStatus,
		"to":      toStatus,
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ($1, 'component_version', $2, $3, $4)
	`, auditAction, entityKey, nullIfEmpty(actor), auditPayload)
	if err != nil {
		return ComponentVersionRow{}, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return ComponentVersionRow{}, err
	}

	row.Status = toStatus
	row.Metadata = meta
	if len(contentRef) > 0 {
		row.ContentRef = contentRef
	}
	return row, nil
}

func (s *Store) requireComponentVersionDraft(ctx context.Context, typ, name, version string) error {
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT cv.status::text
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if status != "draft" {
		return ErrComponentVersionNotDraft
	}
	return nil
}
