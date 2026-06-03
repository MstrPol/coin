package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

const (
	gold  = "\033[33m"
	reset = "\033[0m"
)

// banner выводится только при coin --version.
// Оригинальный шрифт из ascii-art.txt (строки 2-7).
const banner = "\n" +
	gold + " ██████╗ ██████╗ ██╗███╗   ██╗\n" + reset +
	gold + "██╔════╝██╔═══██╗██║████╗  ██║\n" + reset +
	gold + "██║     ██║   ██║██║██╔██╗ ██║\n" + reset +
	gold + "██║     ██║   ██║██║██║╚██╗██║\n" + reset +
	gold + "╚██████╗╚██████╔╝██║██║ ╚████║\n" + reset +
	gold + " ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝\n" + reset

var rootCmd = &cobra.Command{
	Use:     "coin",
	Short:   "Coin CI platform CLI",
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Баннер только при `coin --version`.
	// `coin version` — отдельная команда для CI-скриптов.
	rootCmd.SetVersionTemplate(banner + "  " + gold + "CI Platform" + reset + "  •  {{.Version}}\n\n")

	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(dockerfileCmd)
	rootCmd.AddCommand(rnCmd)
}
