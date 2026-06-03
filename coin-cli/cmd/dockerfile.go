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
		cfg, bundle, err := loadConfigAndBundle(cfgPath)
		if err != nil {
			return err
		}

		if bundle.BuildType() != "container" {
			return fmt.Errorf("template %s/%s build.type=%q — Dockerfile не требуется",
				bundle.Name, bundle.Version, bundle.BuildType())
		}

		outPath, err := dockerfile.Render(cfg, bundle)
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
