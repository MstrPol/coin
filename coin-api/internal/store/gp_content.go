package store

import (
	"context"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/manifest"
)

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
	Capabilities map[string]any       `json:"capabilities"`
	Parameters   []manifest.Parameter `json:"parameters"`
	Pipeline     struct {
		Stages []struct {
			ID    string               `json:"id"`
			Name  string               `json:"name"`
			Steps []manifest.StageStep `json:"steps"`
		} `json:"stages"`
	} `json:"pipeline"`
	Build struct {
		Engine   string                 `json:"engine"`
		Targets  []manifest.BuildTarget `json:"targets"`
		Buildkit struct {
			Targets map[string]string `json:"targets"`
		} `json:"buildkit"`
		Dockerfile struct {
			Path        string `json:"path"`
			ImageTarget string `json:"imageTarget"`
			TestTarget  string `json:"testTarget"`
		} `json:"dockerfile"`
	} `json:"build"`
	ValidateSchema struct {
		ArtifactKey string `json:"artifactKey"`
		SHA256      string `json:"sha256"`
	} `json:"validateSchema"`
	Artifacts struct {
		Containerfiles []struct {
			ID          string `json:"id"`
			ArtifactKey string `json:"artifactKey"`
			SHA256      string `json:"sha256"`
		} `json:"containerfiles"`
	} `json:"artifacts"`
	Deliverables  []manifest.Deliverable `json:"deliverables"`
	Containerfile struct {
		ArtifactKey string `json:"artifactKey"`
		SHA256      string `json:"sha256"`
	} `json:"containerfile"`
}

type gpContentMetadata struct {
	Name          string         `json:"-"`
	Version       string         `json:"-"`
	URL           string         `json:"url"`
	SHA256        string         `json:"sha256"`
	BuildControls map[string]any `json:"buildControls"`
	Capabilities  map[string]any `json:"capabilities"`
}

func contentBundleFromGPContent(meta gpContentMetadata, cref gpContentContentRef) manifest.ContentBundle {
	stages := make([]manifest.TypedStage, 0, len(cref.Pipeline.Stages))
	for _, st := range cref.Pipeline.Stages {
		stages = append(stages, manifest.TypedStage{
			ID:    st.ID,
			Name:  st.Name,
			Steps: st.Steps,
		})
	}
	capabilities := meta.Capabilities
	if len(capabilities) == 0 {
		capabilities = cref.Capabilities
	}
	containerfiles := make([]manifest.NamedContentRef, 0, len(cref.Artifacts.Containerfiles))
	for _, cf := range cref.Artifacts.Containerfiles {
		containerfiles = append(containerfiles, manifest.NamedContentRef{
			ID:          cf.ID,
			ArtifactKey: cf.ArtifactKey,
			SHA256:      cf.SHA256,
		})
	}
	return manifest.ContentBundle{
		Name:                  meta.Name,
		Version:               meta.Version,
		BundleURL:             meta.URL,
		BundleSHA256:          meta.SHA256,
		BuildControls:         meta.BuildControls,
		Capabilities:          capabilities,
		Parameters:            cref.Parameters,
		SchemaArtifactKey:     cref.ValidateSchema.ArtifactKey,
		SchemaSHA256:          cref.ValidateSchema.SHA256,
		ContainerfileKey:      cref.Containerfile.ArtifactKey,
		ContainerfileSHA256:   cref.Containerfile.SHA256,
		Containerfiles:        containerfiles,
		Targets:               cref.Build.Targets,
		Deliverables:          cref.Deliverables,
		BuildEngine:           cref.Build.Engine,
		BuildkitTargets:       cref.Build.Buildkit.Targets,
		DockerfilePath:        cref.Build.Dockerfile.Path,
		DockerfileImageTarget: cref.Build.Dockerfile.ImageTarget,
		DockerfileTestTarget:  cref.Build.Dockerfile.TestTarget,
		Stages:                stages,
	}
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
