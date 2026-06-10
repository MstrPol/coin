package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"coin.local/coin-api/internal/migrate"
	"coin.local/coin-api/internal/scanner"
	"coin.local/coin-api/internal/store"
)

func main() {
	force := flag.Bool("force", false, "rescan all repos ignoring last SHA")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()
	if err := migrate.Up(ctx, sqlDB); err != nil {
		logger.Error("migrations", "err", err)
		os.Exit(1)
	}

	svc := scanner.New(scanner.NewGiteaFromEnv(), store.New(pool), logger)
	result, err := svc.Run(ctx, *force)
	if err != nil {
		logger.Error("scan failed", "err", err)
		os.Exit(1)
	}

	logger.Info("scan complete",
		"total", result.ReposTotal,
		"scanned", result.ReposScanned,
		"skipped", result.ReposSkipped,
		"failed", result.ReposFailed,
		"duration", result.FinishedAt.Sub(result.StartedAt).String(),
	)
	if result.ReposFailed > 0 {
		os.Exit(1)
	}
}
