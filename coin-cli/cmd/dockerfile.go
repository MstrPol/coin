package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/dockerfile"
)

var dockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Managed Dockerfile operations",
}

var dockerfileRenderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render managed Dockerfile into .coin/generated/Dockerfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		if cfg.BuildTarget() != "container" {
			return fmt.Errorf("dockerfile render requires pipeline.build.target: container")
		}

		outPath, err := dockerfile.Render(cfg)
		if err != nil {
			return err
		}

		fmt.Printf("Dockerfile rendered: %s\n", outPath)
		return nil
	},
}

func init() {
	dockerfileRenderCmd.Flags().String("config", config.DefaultPath, "path to .coin/config.yaml")
	dockerfileCmd.AddCommand(dockerfileRenderCmd)
}
