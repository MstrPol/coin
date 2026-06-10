package scanner

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"coin.local/coin-api/internal/metrics"
	"coin.local/coin-api/internal/store"
)

const configPath = ".coin/config.yaml"

type Service struct {
	gitea  *GiteaClient
	store  *store.Store
	logger *slog.Logger
}

func New(gitea *GiteaClient, st *store.Store, logger *slog.Logger) *Service {
	return &Service{gitea: gitea, store: st, logger: logger}
}

func (s *Service) Run(ctx context.Context, force bool) (store.ScanResult, error) {
	start := time.Now()
	result := store.ScanResult{StartedAt: start}
	var runErr error
	defer func() {
		if result.FinishedAt.IsZero() {
			result.FinishedAt = time.Now()
		}
		metrics.ObserveScan(result, runErr)
	}()

	repos, err := s.gitea.ListRepos(ctx)
	if err != nil {
		runErr = err
		return result, err
	}
	result.ReposTotal = len(repos)

	for _, repo := range repos {
		if err := ctx.Err(); err != nil {
			runErr = err
			return result, err
		}
		owner, name, err := SplitOwnerRepo(repo.FullName)
		if err != nil {
			result.ReposFailed++
			s.logger.Warn("skip repo", "repo", repo.FullName, "err", err)
			continue
		}

		sha, err := s.gitea.BranchSHA(ctx, owner, name, repo.DefaultBranch)
		if err != nil {
			result.ReposFailed++
			s.logger.Warn("branch sha", "repo", repo.FullName, "err", err)
			continue
		}

		if !force {
			prev, ok, err := s.store.ScannerLastSHA(ctx, repo.FullName)
			if err != nil {
				runErr = err
				return result, err
			}
			if ok && prev == sha {
				result.ReposSkipped++
				continue
			}
		}

		raw, err := s.gitea.RawFile(ctx, owner, name, repo.DefaultBranch, configPath)
		if errors.Is(err, ErrNoConfig) {
			result.ReposSkipped++
			_ = s.store.SaveScannerSHA(ctx, repo.FullName, sha)
			continue
		}
		if err != nil {
			result.ReposFailed++
			s.logger.Warn("fetch config", "repo", repo.FullName, "err", err)
			continue
		}

		parsed, err := ParseConfig(raw, repo.Name)
		if errors.Is(err, ErrConfigV1) {
			result.ReposSkipped++
			s.logger.Info("skip v1 config", "repo", repo.FullName)
			_ = s.store.SaveScannerSHA(ctx, repo.FullName, sha)
			continue
		}
		if err != nil {
			result.ReposFailed++
			s.logger.Warn("parse config", "repo", repo.FullName, "err", err)
			continue
		}

		gitURL := s.gitea.CloneURL(repo.FullName)
		if err := s.store.UpsertProjectScan(ctx, store.ProjectScanInput{
			Project:    parsed.Project,
			GoldenPath: parsed.GoldenPath,
			Version:    parsed.Version,
			GitURL:     gitURL,
		}); err != nil {
			result.ReposFailed++
			s.logger.Warn("upsert project", "repo", repo.FullName, "err", err)
			continue
		}

		if err := s.store.SaveScannerSHA(ctx, repo.FullName, sha); err != nil {
			runErr = err
			return result, err
		}
		result.ReposScanned++
		s.logger.Info("scanned", "repo", repo.FullName, "project", parsed.Project,
			"gp", parsed.GoldenPath, "version", parsed.Version)
	}

	result.FinishedAt = time.Now()
	return result, nil
}
