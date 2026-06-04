package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"coin.local/coin-cli/internal/platform"
)

var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Операции с coin-platform",
}

var platformValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Проверить связность golden-paths, starters и agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := platform.Validate(); err != nil {
			return err
		}
		root, _ := platform.Root()
		fmt.Fprintf(os.Stdout, "✓ platform OK (%s)\n", root)
		return nil
	},
}

func init() {
	platformCmd.AddCommand(platformValidateCmd)
	rootCmd.AddCommand(platformCmd)
}
