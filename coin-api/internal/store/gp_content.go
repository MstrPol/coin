package store

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/manifest"
)

// CanonicalGPSlots returns the five platform composition slots for a golden path.
func CanonicalGPSlots(gpName, agentStack string) []GPProfileSlot {
	return []GPProfileSlot{
		{Key: "jnlp", Type: "agent", Name: "jnlp"},
		{Key: "agent", Type: "agent", Name: agentStack},
		{Key: "executor", Type: "executor", Name: "coin-executor"},
		{Key: "lib", Type: "lib", Name: "coin-lib"},
		{Key: "gp-content", Type: "gp-content", Name: gpName},
	}
}

type gpContentContentRef struct {
	Stages []struct {
		Name        string `json:"name"`
		When        string `json:"when"`
		ArtifactKey string `json:"artifactKey"`
		SHA256      string `json:"sha256"`
	} `json:"stages"`
	ValidateSchema struct {
		ArtifactKey string `json:"artifactKey"`
		SHA256      string `json:"sha256"`
	} `json:"validateSchema"`
	DockerfileTemplate struct {
		ArtifactKey string `json:"artifactKey"`
		SHA256      string `json:"sha256"`
	} `json:"dockerfileTemplate"`
}

type gpContentMetadata struct {
	URL           string         `json:"url"`
	SHA256        string         `json:"sha256"`
	BuildControls map[string]any `json:"buildControls"`
	Capabilities  map[string]any `json:"capabilities"`
}

func (s *Store) getGPContentRefs(ctx context.Context, name, version string) (gpContentMetadata, gpContentContentRef, error) {
	var metaRaw, crefRaw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT cv.metadata, COALESCE(cv.content_ref, 'null'::jsonb)
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = 'gp-content' AND c.name = $1 AND cv.version = $2 AND cv.status = 'published'
	`, name, version).Scan(&metaRaw, &crefRaw)
	if err == pgx.ErrNoRows {
		return gpContentMetadata{}, gpContentContentRef{}, fmt.Errorf("gp-content/%s@%s not found", name, version)
	}
	if err != nil {
		return gpContentMetadata{}, gpContentContentRef{}, err
	}
	var meta gpContentMetadata
	if err := json.Unmarshal(metaRaw, &meta); err != nil {
		return gpContentMetadata{}, gpContentContentRef{}, fmt.Errorf("gp-content metadata: %w", err)
	}
	var cref gpContentContentRef
	if err := json.Unmarshal(crefRaw, &cref); err != nil {
		return gpContentMetadata{}, gpContentContentRef{}, fmt.Errorf("gp-content content_ref: %w", err)
	}
	return meta, cref, nil
}

func contentBundleFromGPContent(meta gpContentMetadata, cref gpContentContentRef) manifest.ContentBundle {
	stages := make([]manifest.StageScript, 0, len(cref.Stages))
	for _, st := range cref.Stages {
		stages = append(stages, manifest.StageScript{
			Name:        st.Name,
			When:        st.When,
			ArtifactKey: st.ArtifactKey,
			SHA256:      st.SHA256,
		})
	}
	return manifest.ContentBundle{
		BundleURL:             meta.URL,
		BundleSHA256:          meta.SHA256,
		BuildControls:         meta.BuildControls,
		Capabilities:          meta.Capabilities,
		SchemaArtifactKey:     cref.ValidateSchema.ArtifactKey,
		SchemaSHA256:          cref.ValidateSchema.SHA256,
		DockerfileArtifactKey: cref.DockerfileTemplate.ArtifactKey,
		DockerfileSHA256:      cref.DockerfileTemplate.SHA256,
		Stages:                stages,
	}
}

func (s *Store) seedGPArtifactsFromGPContent(ctx context.Context, releaseID int64, name, version string) error {
	_, cref, err := s.getGPContentRefs(ctx, name, version)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(cref.Stages)+2)
	for _, st := range cref.Stages {
		keys = append(keys, st.ArtifactKey)
	}
	keys = append(keys, cref.ValidateSchema.ArtifactKey, cref.DockerfileTemplate.ArtifactKey)

	var componentVersionID int64
	err = s.pool.QueryRow(ctx, `
		SELECT cv.id FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = 'gp-content' AND c.name = $1 AND cv.version = $2
	`, name, version).Scan(&componentVersionID)
	if err != nil {
		return err
	}

	for _, key := range keys {
		body, sha, err := s.loadComponentArtifactBody(ctx, componentVersionID, key)
		if err == pgx.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}
		_, err = s.pool.Exec(ctx, `
			INSERT INTO gp_artifact_bodies (gp_release_id, artifact_key, body, sha256)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (gp_release_id, artifact_key) DO UPDATE SET body = EXCLUDED.body, sha256 = EXCLUDED.sha256
		`, releaseID, key, body, sha)
		if err != nil {
			return fmt.Errorf("seed gp artifact %s: %w", key, err)
		}
	}
	return nil
}

func (s *Store) loadComponentArtifactBody(ctx context.Context, componentVersionID int64, key string) ([]byte, string, error) {
	candidates := []string{key}
	if enc := strings.ReplaceAll(key, "/", "%2F"); enc != key {
		candidates = append(candidates, enc)
	}
	if dec, err := url.PathUnescape(key); err == nil && dec != key {
		candidates = append(candidates, dec)
	}
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		var body []byte
		var sha string
		err := s.pool.QueryRow(ctx, `
			SELECT body, sha256 FROM component_artifact_bodies
			WHERE component_version_id = $1 AND artifact_key = $2
		`, componentVersionID, candidate).Scan(&body, &sha)
		if err == nil {
			return body, sha, nil
		}
		if err != pgx.ErrNoRows {
			return nil, "", err
		}
	}
	return nil, "", pgx.ErrNoRows
}

func (s *Store) gpContentVersionFromComposition(ctx context.Context, gpName, gpVersion string) (string, string, error) {
	var name, ver string
	err := s.pool.QueryRow(ctx, `
		SELECT gc.component_name, gc.component_version
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2
		  AND gc.component_type = 'gp-content'
	`, gpName, gpVersion).Scan(&name, &ver)
	if err == pgx.ErrNoRows {
		return "", "", fmt.Errorf("gp-content not in composition for %s@%s", gpName, gpVersion)
	}
	return name, ver, err
}

func (s *Store) LibVersionFromComposition(ctx context.Context, gpName, gpVersion string) (string, string, error) {
	var name, ver string
	err := s.pool.QueryRow(ctx, `
		SELECT gc.component_name, gc.component_version
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2
		  AND gc.component_type = 'lib'
	`, gpName, gpVersion).Scan(&name, &ver)
	if err == pgx.ErrNoRows {
		return "", "", fmt.Errorf("lib not in composition for %s@%s", gpName, gpVersion)
	}
	return name, ver, err
}

func (s *Store) SaveComponentArtifactBody(ctx context.Context, typ, name, version, artifactKey string, body []byte, sha256sum string) error {
	var versionID int64
	err := s.pool.QueryRow(ctx, `
		SELECT cv.id FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = $1 AND c.name = $2 AND cv.version = $3
	`, typ, name, version).Scan(&versionID)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO component_artifact_bodies (component_version_id, artifact_key, body, sha256)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (component_version_id, artifact_key)
		DO UPDATE SET body = EXCLUDED.body, sha256 = EXCLUDED.sha256
	`, versionID, artifactKey, body, sha256sum)
	return err
}
