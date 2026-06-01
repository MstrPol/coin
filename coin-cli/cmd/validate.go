package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"coin.local/coin-cli/internal/config"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate .coin/config.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		minCLI, _ := cmd.Flags().GetString("min-version")
		if minCLI != "" && Version != "dev" {
			// простая лексическая проверка для semver X.Y.Z
			if Version < minCLI {
				return fmt.Errorf(
					"coin CLI %s ниже минимально допустимой версии %s. Обновите образ агента.",
					Version, minCLI,
				)
			}
		}

		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		fmt.Printf("✓ config valid: project=%s stack=%s\n", cfg.Project.Name, cfg.Project.Stack)
		return nil
	},
}

func init() {
	validateCmd.Flags().String("config", config.DefaultPath, "path to .coin/config.yaml")
	validateCmd.Flags().String("min-version", "", "minimum coin CLI version required")
}
