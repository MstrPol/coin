package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/executor"
)

var runCmd = &cobra.Command{
	Use:   "run <test|build|publish>",
	Short: "Run a CI stage (test, build, publish)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stage := args[0]
		if stage != "test" && stage != "build" && stage != "publish" {
			return fmt.Errorf("unknown stage %q: must be test, build or publish", stage)
		}

		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, bundle, err := loadConfigAndBundle(cfgPath)
		if err != nil {
			return err
		}

		return executor.New(cfg, bundle).Run(stage)
	},
}

func init() {
	runCmd.Flags().String("config", config.DefaultPath, "path to .coin/config.yaml")
}
