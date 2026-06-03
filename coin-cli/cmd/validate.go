package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/goldenpaths"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate .coin/config.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		minCLI, _ := cmd.Flags().GetString("min-version")
		if minCLI != "" && Version != "dev" {
			if Version < minCLI {
				return fmt.Errorf(
					"coin CLI %s ниже минимально допустимой версии %s. Обновите CLI из Nexus.",
					Version, minCLI,
				)
			}
		}

		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, bundle, err := loadConfigAndBundle(cfgPath)
		if err != nil {
			return err
		}

		stack := bundle.Stack()
		if cfg.Jenkins.Stack != "" {
			stack = cfg.Jenkins.Stack
		}

		fmt.Printf("✓ config valid: project=%s template=%s/%s stack=%s\n",
			cfg.Project.Name, bundle.Name, bundle.Version, stack)

		if entry, ok := bundle.Catalog.Paths[bundle.Name]; ok && entry.Latest != "" && entry.Latest != bundle.Version {
			fmt.Printf("ℹ latest template version: %s (current %s)\n", entry.Latest, bundle.Version)
		}

		if msg := bundle.Catalog.DeprecationWarning(bundle.Name, bundle.Version); msg != "" {
			fmt.Printf("⚠ deprecated: %s\n", msg)
		}

		fmt.Printf("ℹ golden paths source: %s\n", goldenpaths.SourceLabel())

		return nil
	},
}

func init() {
	validateCmd.Flags().String("config", config.DefaultPath, "path to .coin/config.yaml")
	validateCmd.Flags().String("min-version", "", "minimum coin CLI version required")
}
