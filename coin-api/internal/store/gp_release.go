package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/compatibility"
	"coin.local/coin-api/internal/gpcontent"
	"coin.local/coin-api/internal/manifest"
	"coin.local/coin-api/internal/pin"
)

var (
	ErrDuplicateGPRelease   = errors.New("gp release already exists")
	ErrComponentNotFound    = errors.New("component version not found")
	ErrInvalidComposition   = errors.New("invalid composition")
)

type PublishGPReleaseInput struct {
	Name        string
	Version     string
	Composition map[string]string
	Actor       string
}

type GPReleaseRow struct {
	ID           int64
	Name         string
	Version      string
	GitExportTag string
	Status       string
}

func (s *Store) PublishGPRelease(ctx context.Context, in PublishGPReleaseInput) (GPReleaseRow, error) {
	if in.Name == "" || in.Version == "" {
		return GPReleaseRow{}, fmt.Errorf("name and version are required")
	}
	if pin.IsSnapshotVersion(in.Version) {
		return GPReleaseRow{}, fmt.Errorf("published version cannot contain snapshot suffix")
	}

	slots, err := s.profileSlots(ctx, in.Name)
	if err != nil {
		return GPReleaseRow{}, err
	}

	rules, err := s.loadCompatibilityRules(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	if err := compatibility.Validate(slots, in.Composition, rules); err != nil {
		return GPReleaseRow{}, fmt.Errorf("%w: %v", ErrInvalidComposition, err)
	}

	for _, slot := range slots {
		ver := in.Composition[slot.Key]
		ok, err := s.componentVersionResolvable(ctx, slot.Type, slot.Name, ver, ComponentResolveStable)
		if err != nil {
			return GPReleaseRow{}, err
		}
		if !ok {
			return GPReleaseRow{}, fmt.Errorf("%w: %s/%s@%s", ErrComponentNotFound, slot.Type, slot.Name, ver)
		}
	}

	var releaseID int64
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO gp_releases (name, version, status)
		VALUES ($1, $2, 'published')
		RETURNING id
	`, in.Name, in.Version).Scan(&releaseID)
	if err != nil {
		if isUniqueViolation(err) {
			return GPReleaseRow{}, ErrDuplicateGPRelease
		}
		return GPReleaseRow{}, fmt.Errorf("gp release insert: %w", err)
	}

	for _, slot := range slots {
		_, err = tx.Exec(ctx, `
			INSERT INTO gp_composition (gp_release_id, component_type, component_name, component_version)
			VALUES ($1, $2, $3, $4)
		`, releaseID, slot.Type, slot.Name, in.Composition[slot.Key])
		if err != nil {
			return GPReleaseRow{}, fmt.Errorf("composition insert: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO catalog_policy (gp_name, latest, minimum, deprecated)
		VALUES ($1, $2, $2, '[]'::jsonb)
		ON CONFLICT (gp_name) DO UPDATE SET latest = EXCLUDED.latest
	`, in.Name, in.Version)
	if err != nil {
		return GPReleaseRow{}, fmt.Errorf("catalog policy: %w", err)
	}

	entityKey := fmt.Sprintf("%s@%s", in.Name, in.Version)
	auditPayload, _ := json.Marshal(map[string]any{
		"name":        in.Name,
		"version":     in.Version,
		"composition": in.Composition,
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('publish_gp_release', 'gp_release', $1, $2, $3)
	`, entityKey, nullIfEmpty(in.Actor), auditPayload)
	if err != nil {
		return GPReleaseRow{}, fmt.Errorf("audit log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return GPReleaseRow{}, err
	}

	seeded := false
	if contentName, contentVer, err := s.gpContentVersionFromComposition(ctx, in.Name, in.Version); err == nil {
		if err := s.seedGPArtifactsFromGPContent(ctx, releaseID, contentName, contentVer); err == nil {
			seeded = true
		}
	}
	if !seeded {
		_ = gpcontent.SeedArtifactsToRelease(ctx, s.pool, in.Name, releaseID)
	}

	return GPReleaseRow{
		ID:      releaseID,
		Name:    in.Name,
		Version: in.Version,
		Status:  "published",
	}, nil
}

func (s *Store) componentVersionResolvable(ctx context.Context, typ, name, version string, mode ComponentResolveMode) (bool, error) {
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT cv.status::text
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version).Scan(&status)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return componentStatusAllowed(status, mode), nil
}

// componentVersionPublished is stable-channel resolve (published only).
func (s *Store) componentVersionPublished(ctx context.Context, typ, name, version string) (bool, error) {
	return s.componentVersionResolvable(ctx, typ, name, version, ComponentResolveStable)
}

func (s *Store) loadCompatibilityRules(ctx context.Context) ([]compatibility.Rule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT source_type, source_name, source_version_prefix, requirements
		FROM component_compatibility
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []compatibility.Rule
	for rows.Next() {
		var rule compatibility.Rule
		var raw []byte
		if err := rows.Scan(&rule.SourceType, &rule.SourceName, &rule.VersionPrefix, &raw); err != nil {
			return nil, err
		}
		reqs, err := compatibility.ParseRequirements(raw)
		if err != nil {
			return nil, err
		}
		rule.Requirements = reqs
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (s *Store) loadContentBundle(ctx context.Context, gpName, gpVersion string, mode ComponentResolveMode) (manifest.ContentBundle, error) {
	contentName, contentVer, err := s.gpContentVersionFromComposition(ctx, gpName, gpVersion)
	if err != nil {
		return manifest.ContentBundle{}, err
	}
	return s.materializeGPContentBundle(ctx, contentName, contentVer, mode)
}

func (s *Store) materializeGPContentBundle(ctx context.Context, name, version string, mode ComponentResolveMode) (manifest.ContentBundle, error) {
	meta, crefRaw, err := s.getGPContentRefs(ctx, name, version, mode)
	if err != nil {
		return manifest.ContentBundle{}, err
	}
	return contentBundleFromRawRef(meta, crefRaw)
}
