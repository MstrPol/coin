package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"

	"coin.local/coin-api/internal/gpcontent"
	"coin.local/coin-api/migrations"
)

func Up(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	if err := gpcontent.SeedGoAppV100(ctx, db); err != nil {
		return fmt.Errorf("gpcontent seed: %w", err)
	}
	return nil
}

// MustFS exposes migrations for tests.
func MustFS() fs.FS {
	return migrations.FS
}
