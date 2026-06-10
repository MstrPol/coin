package gpcontent

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
)

//go:embed seed/**
var seedFS embed.FS

type seedArtifact struct {
	key     string
	relPath string
}

var goAppV100 = []seedArtifact{
	{key: "scripts/validate.sh", relPath: "seed/go-app/1.0.0/scripts/validate.sh"},
	{key: "scripts/test.sh", relPath: "seed/go-app/1.0.0/scripts/test.sh"},
	{key: "scripts/build.sh", relPath: "seed/go-app/1.0.0/scripts/build.sh"},
	{key: "scripts/publish.sh", relPath: "seed/go-app/1.0.0/scripts/publish.sh"},
	{key: "Dockerfile", relPath: "seed/go-app/1.0.0/Dockerfile"},
	{key: "schema/config.v2.schema.json", relPath: "seed/schema/config.v2.schema.json"},
	{key: "orchestration/coinPipeline.groovy", relPath: "seed/orchestration/coinPipeline.groovy"},
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
