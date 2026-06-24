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
func CanonicalGPSlots(gpName string) []GPProfileSlot {
	return []GPProfileSlot{
		{Key: "agent", Type: "agent", Name: "coin-agent"},
		{Key: "executor", Type: "executor", Name: "coin-executor"},
		{Key: "lib", Type: "lib", Name: "coin-lib"},
		{Key: "gp-content", Type: "gp-content", Name: gpName},
		{Key: "branching-model", Type: "branching-model", Name: DefaultBranchingModelForGP(gpName)},
	}
}

// DefaultBranchingModelForGP picks the reference branching model name for a GP profile.
func DefaultBranchingModelForGP(gpName string) string {
	switch gpName {
	case "go-lib", "java-maven-app":
		return "semver-tag"
	default:
		return "trunk-based"
	}
}

type gpContentContentRef struct {
	Pipeline struct {
		Stages []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			When string `json:"when"`
		} `json:"stages"`
	} `json:"pipeline"`
	Build struct {
		Engine   string `json:"engine"`
		Buildkit struct {
			Dockerfile       string            `json:"dockerfile"`
			Targets          map[string]string `json:"targets"`
			CacheRefTemplate string            `json:"cacheRefTemplate"`
		} `json:"buildkit"`
		Buildpack struct {
			Builder          string `json:"builder"`
			RunImage         string `json:"runImage"`
			CacheRefTemplate string `json:"cacheRefTemplate"`
		} `json:"buildpack"`
		Dockerfile struct {
			File             string `json:"file"`
			ImageTarget      string `json:"imageTarget"`
			TestTarget       string `json:"testTarget"`
			CacheRefTemplate string `json:"cacheRefTemplate"`
		} `json:"dockerfile"`
	} `json:"build"`
	ValidateSchema struct {
		ArtifactKey string `json:"artifactKey"`
		SHA256      string `json:"sha256"`
	} `json:"validateSchema"`
	Containerfile struct {
		ArtifactKey string `json:"artifactKey"`
		SHA256      string `json:"sha256"`
	} `json:"containerfile"`
}

type gpContentMetadata struct {
	URL           string         `json:"url"`
	SHA256        string         `json:"sha256"`
	BuildControls map[string]any `json:"buildControls"`
	Capabilities  map[string]any `json:"capabilities"`
}

func (s *Store) getGPContentRefs(ctx context.Context, name, version string, mode ComponentResolveMode) (gpContentMetadata, json.RawMessage, error) {
	allowed := allowedComponentStatuses(mode)
	var metaRaw, crefRaw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT cv.metadata, COALESCE(cv.content_ref, 'null'::jsonb)
		FROM component_versions cv
		JOIN components c ON c.id = cv.component_id
		WHERE c.type = 'gp-content' AND c.name = $1 AND cv.version = $2
		  AND cv.status::text = ANY($3)
	`, name, version, allowed).Scan(&metaRaw, &crefRaw)
	if err == pgx.ErrNoRows {
		return gpContentMetadata{}, nil, fmt.Errorf("gp-content/%s@%s not found or not visible for resolve mode %q", name, version, mode)
	}
	if err != nil {
		return gpContentMetadata{}, nil, err
	}
	var meta gpContentMetadata
	if err := json.Unmarshal(metaRaw, &meta); err != nil {
		return gpContentMetadata{}, nil, fmt.Errorf("gp-content metadata: %w", err)
	}
	return meta, json.RawMessage(crefRaw), nil
}

func contentBundleFromGPContent(meta gpContentMetadata, cref gpContentContentRef) manifest.ContentBundle {
	stages := make([]manifest.TypedStage, 0, len(cref.Pipeline.Stages))
	for _, st := range cref.Pipeline.Stages {
		stages = append(stages, manifest.TypedStage{
			ID:   st.ID,
			Name: st.Name,
			When: st.When,
		})
	}
	return manifest.ContentBundle{
		BundleURL:           meta.URL,
		BundleSHA256:        meta.SHA256,
		BuildControls:       meta.BuildControls,
		Capabilities:        meta.Capabilities,
		SchemaArtifactKey:   cref.ValidateSchema.ArtifactKey,
		SchemaSHA256:        cref.ValidateSchema.SHA256,
		ContainerfileKey:    cref.Containerfile.ArtifactKey,
		ContainerfileSHA256: cref.Containerfile.SHA256,
		BuildEngine:         cref.Build.Engine,
		BuildkitDockerfile:  cref.Build.Buildkit.Dockerfile,
		BuildkitTargets:     cref.Build.Buildkit.Targets,
		BuildpackBuilder:      cref.Build.Buildpack.Builder,
		BuildpackRunImage:     cref.Build.Buildpack.RunImage,
		DockerfileImageTarget: cref.Build.Dockerfile.ImageTarget,
		DockerfileTestTarget:  cref.Build.Dockerfile.TestTarget,
		CacheRefTemplate:    cacheRefTemplateFromContent(cref),
		Stages:              stages,
	}
}

func cacheRefTemplateFromContent(cref gpContentContentRef) string {
	if t := strings.TrimSpace(cref.Build.Buildkit.CacheRefTemplate); t != "" {
		return t
	}
	if t := strings.TrimSpace(cref.Build.Buildpack.CacheRefTemplate); t != "" {
		return t
	}
	return strings.TrimSpace(cref.Build.Dockerfile.CacheRefTemplate)
}

func (s *Store) seedGPArtifactsFromGPContent(ctx context.Context, releaseID int64, name, version string) error {
	_, crefRaw, err := s.getGPContentRefs(ctx, name, version, ComponentResolveAdmin)
	if err != nil {
		return err
	}
	var cref gpContentContentRef
	if isContentRefV2(crefRaw) {
		bundle, err := contentBundleFromV2Manifest(gpContentMetadata{}, crefRaw)
		if err != nil {
			return err
		}
		cref.ValidateSchema.ArtifactKey = bundle.SchemaArtifactKey
		cref.ValidateSchema.SHA256 = bundle.SchemaSHA256
		cref.Containerfile.ArtifactKey = bundle.ContainerfileKey
		cref.Containerfile.SHA256 = bundle.ContainerfileSHA256
	} else if err := json.Unmarshal(crefRaw, &cref); err != nil {
		return fmt.Errorf("gp-content content_ref: %w", err)
	}
	keys := []string{cref.ValidateSchema.ArtifactKey}
	if cf := strings.TrimSpace(cref.Containerfile.ArtifactKey); cf != "" {
		keys = append(keys, cf)
	}

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
	if err := s.requireComponentVersionDraft(ctx, typ, name, version); err != nil {
		return err
	}
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
