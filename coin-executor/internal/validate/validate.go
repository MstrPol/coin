package validate

import (
	"fmt"
	"os"

	"coin.local/coin-executor/pkg/branching"
	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/deliverables"
	"coin.local/coin-executor/internal/manifest"
	"coin.local/coin-executor/internal/policy"
)

func Project(cfg *config.Config, m *manifest.Manifest, workDir string) error {
	if err := m.MatchesConfig(cfg.Coin.GoldenPath, cfg.Coin.Version); err != nil {
		return err
	}
	if model := branching.FromManifest(m); model != nil {
		g, err := branching.GitFromEnv(workDir)
		if err != nil {
			return fmt.Errorf("git context: %w", err)
		}
		if err := branching.ValidateBranch(model, g.Branch); err != nil {
			return fmt.Errorf("branching: %w", err)
		}
		version, err := branching.ResolveVersion(model, g)
		if err != nil {
			return fmt.Errorf("branching version: %w", err)
		}
		fmt.Printf("✓ branching: branch=%s coin_version=%s model=%s@%s\n",
			g.Branch, version, model.Name, model.Version)
	}
	resolved := m.GoldenPath.Version
	if resolved == "" {
		return fmt.Errorf("manifest missing goldenPath.version")
	}
	check, err := policy.CheckResolvedVersion(cfg.Coin.GoldenPath, resolved)
	if err != nil {
		return err
	}
	if check.Warning != "" {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", check.Warning)
	}
	items := cfg.NormalizedDeliverables()
	if err := deliverables.Validate(items, m.AllowedDeliverableTypes()); err != nil {
		return err
	}
	fmt.Printf("✓ config valid: project=%s gp=%s pin=%s resolved=%s deliverables=%d\n",
		cfg.Project.Name, cfg.Coin.GoldenPath, cfg.Coin.Version, resolved, len(items))
	return nil
}
