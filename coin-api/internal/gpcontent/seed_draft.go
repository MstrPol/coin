package gpcontent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedArtifactsToRelease copies embedded go-app artifact bodies onto a draft release.
// DEPRECATED (GCP-5): bootstrap fallback when gp-content component is not registered.
// Prefer seed-jenkins-lib or Component Studio register — see docs/runbooks/gp-artifact-bodies-migration.md.
func SeedArtifactsToRelease(ctx context.Context, pool *pgxpool.Pool, gpName string, releaseID int64) error {
	if gpName != "go-app" {
		return nil
	}
	for _, a := range goAppV100 {
		body, err := seedFS.ReadFile(a.relPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", a.relPath, err)
		}
		sum := sha256.Sum256(body)
		hash := "sha256:" + hex.EncodeToString(sum[:])
		_, err = pool.Exec(ctx, `
			INSERT INTO gp_artifact_bodies (gp_release_id, artifact_key, body, sha256)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (gp_release_id, artifact_key) DO NOTHING
		`, releaseID, a.key, body, hash)
		if err != nil {
			return fmt.Errorf("insert artifact %s: %w", a.key, err)
		}
	}
	return nil
}
