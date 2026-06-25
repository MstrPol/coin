package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/compatibility"
	"coin.local/coin-api/internal/componentpackage"
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
	Name               string
	Version            string
	Composition        map[string]string
	AgentStackName     string
	GPContentName      string
	BranchingModelName string
	Actor              string
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

	prep, err := s.prepareGPRelease(ctx, in)
	if err != nil {
		return GPReleaseRow{}, err
	}

	rules, err := s.loadCompatibilityRules(ctx)
	if err != nil {
		return GPReleaseRow{}, err
	}
	if err := validateGPReleaseComposition(s, ctx, prep, rules, componentResolveModeForGPPublish); err != nil {
		return GPReleaseRow{}, err
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

	if err := insertGPComposition(ctx, tx, releaseID, prep.storeSlots, in.Composition); err != nil {
		return GPReleaseRow{}, err
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

// componentResolveModeForGPPublish selects which component statuses are valid when pinning a GP release.
// branching-model may be canary-only until promote (BML); other slots require published for stable GP lines.
func componentResolveModeForGPPublish(componentType string) ComponentResolveMode {
	if componentpackage.UsesPGOnlyCanaryRegistry(componentType) {
		return ComponentResolveCanary
	}
	return ComponentResolveStable
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
