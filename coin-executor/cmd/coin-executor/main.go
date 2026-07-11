package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"coin.local/coin-executor/pkg/branching"
	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/executor"
	"coin.local/coin-executor/internal/manifest"
	"coin.local/coin-executor/internal/report"
	"coin.local/coin-executor/internal/validate"
)

var (
	Version = "0.1.0-dev"
)

func main() {
	root := &cobra.Command{
		Use:   "coin-executor",
		Short: "Coin runtime executor",
	}

	root.AddCommand(validateCmd())
	root.AddCommand(runCmd())
	root.AddCommand(versionCmd())
	root.AddCommand(reportCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func validateCmd() *cobra.Command {
	var projectPath, manifestPath string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate project config against manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(projectPath)
			if err != nil {
				return err
			}
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			return validate.Project(cfg, m, wd)
		},
	}
	cmd.Flags().StringVar(&projectPath, "project", config.DefaultPath, "project config path")
	cmd.Flags().StringVar(&manifestPath, "manifest", ".coin/manifest.json", "manifest path")
	return cmd
}

func runCmd() *cobra.Command {
	var projectPath, manifestPath, stage string
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run pipeline stages from manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(projectPath)
			if err != nil {
				return err
			}
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}
			if err := m.MatchesConfig(cfg.Coin.GoldenPath, cfg.Coin.Version); err != nil {
				return err
			}
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			return executor.Runner{Workspace: wd}.Run(cfg, m, executor.RunOptions{Stage: stage})
		},
	}
	cmd.Flags().StringVar(&projectPath, "project", config.DefaultPath, "project config path")
	cmd.Flags().StringVar(&manifestPath, "manifest", ".coin/manifest.json", "manifest path")
	cmd.Flags().StringVar(&stage, "stage", "", "run single stage (validate|test|build|publish)")
	return cmd
}

func versionCmd() *cobra.Command {
	var manifestPath string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print product version (COIN_VERSION) or executor binary version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if m, err := manifest.Load(manifestPath); err == nil {
				if model := branching.FromManifest(m); model != nil {
					wd, err := os.Getwd()
					if err != nil {
						return err
					}
					g, err := branching.GitFromEnv(wd)
					if err != nil {
						return err
					}
					v, err := branching.ResolveVersion(model, g)
					if err != nil {
						return err
					}
					fmt.Println(v)
					return nil
				}
			}
			fmt.Println(Version)
			return nil
		},
	}
	cmd.Flags().StringVar(&manifestPath, "manifest", ".coin/manifest.json", "manifest path")
	return cmd
}

func reportCmd() *cobra.Command {
	var manifestPath, projectPath, buildURL, result string
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report build telemetry to coin-api",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := report.Submit(projectPath, manifestPath, buildURL, result); err != nil {
				return err
			}
			fmt.Println("✓ build report sent")
			return nil
		},
	}
	cmd.Flags().StringVar(&manifestPath, "manifest", ".coin/manifest.json", "manifest path")
	cmd.Flags().StringVar(&projectPath, "project", config.DefaultPath, "project config path")
	cmd.Flags().StringVar(&buildURL, "build-url", "", "jenkins build URL")
	cmd.Flags().StringVar(&result, "result", "success", "build result")
	return cmd
}
