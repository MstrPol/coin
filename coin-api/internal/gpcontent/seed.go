// Package gpcontent provides legacy embedded seed artifacts for local bootstrap.
// DEPRECATED (GCP-5): see docs/runbooks/gp-artifact-bodies-migration.md Phase C.
// SoT for gp-content is Nexus package + Component Studio; embedded bytes are bootstrap-only.
package gpcontent

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
)

//go:embed seed/schema/config.v2.schema.json seed/go-app/1.0.0/dockerfiles/Containerfile
var seedFS embed.FS

// ValidateSchemaBytes returns embedded config schema for GP release seeding.
func ValidateSchemaBytes() ([]byte, error) {
	return seedFS.ReadFile("seed/schema/config.v2.schema.json")
}

type seedArtifact struct {
	key     string
	relPath string
}

// Embedded fallback artifacts for go-app@1.0.0 draft seeding (gp-content publish is SoT).
var goAppV100 = []seedArtifact{
	{key: "schemas/config.v2.schema.json", relPath: "seed/schema/config.v2.schema.json"},
	{key: "dockerfiles/Containerfile", relPath: "seed/go-app/1.0.0/dockerfiles/Containerfile"},
}

// SeedGoAppV100 inserts embedded artifact bytes for go-app@1.0.0 (idempotent).
func SeedGoAppV100(ctx context.Context, db *sql.DB) error {
	var releaseID int64
	err := db.QueryRowContext(ctx, `
		SELECT id FROM gp_releases WHERE name='go-app' AND version='1.0.0'
	`).Scan(&releaseID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("gp release lookup: %w", err)
	}

	for _, a := range goAppV100 {
		body, err := seedFS.ReadFile(a.relPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", a.relPath, err)
		}
		sum := sha256.Sum256(body)
		hash := "sha256:" + hex.EncodeToString(sum[:])

		_, err = db.ExecContext(ctx, `
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
