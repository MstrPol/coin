package cmd

import (
	"fmt"

	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/goldenpaths"
)

func loadConfigAndBundle(cfgPath string) (*config.Config, *goldenpaths.Bundle, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, nil, err
	}
	bundle, err := goldenpaths.Resolve(cfg.Coin.Template, cfg.Coin.TemplateVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("template: %w", err)
	}
	return cfg, bundle, nil
}
