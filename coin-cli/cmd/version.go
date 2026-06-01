package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/versioning"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Compute COIN_VERSION from git",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		result, err := versioning.Compute(cfg.TagPrefix())
		if err != nil {
			return err
		}

		fmt.Printf("COIN_VERSION=%s\n", result.Version)
		fmt.Printf("COIN_IMAGE_TAG=%s\n", result.ImageTag)
		fmt.Printf("COIN_VERSION_SOURCE=%s\n", result.Source)
		return nil
	},
}

func init() {
	versionCmd.Flags().String("config", config.DefaultPath, "path to .coin/config.yaml")
}
