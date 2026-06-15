package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"coin.local/coin-executor/internal/bootstrap"
	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/deliverables"
	"coin.local/coin-executor/internal/executor"
	"coin.local/coin-executor/internal/manifest"
	"coin.local/coin-executor/internal/policy"
	"coin.local/coin-executor/internal/report"
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
	root.AddCommand(bootstrapCmd())
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
			if err := m.MatchesConfig(cfg.Coin.GoldenPath, cfg.Coin.Version); err != nil {
				return err
			}
			resolved := m.GoldenPath.Version
			if resolved == "" {
				return fmt.Errorf("manifest missing goldenPath.version")
			}
			check, err := policy.CheckResolvedVersion(cfg.Coin.GoldenPath, resolved)
			if err != nil {
				return err
			}
			if check.Warning != "" {
				fmt.Fprintf(os.Stderr, "WARNING: %s\n", check.Warning)
			}
			items := cfg.NormalizedDeliverables()
			if err := deliverables.Validate(items, m.AllowedDeliverableTypes()); err != nil {
				return err
			}
			fmt.Printf("✓ config valid: project=%s gp=%s pin=%s resolved=%s deliverables=%d\n",
				cfg.Project.Name, cfg.Coin.GoldenPath, cfg.Coin.Version, resolved, len(items))
			return nil
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

func bootstrapCmd() *cobra.Command {
	var manifestPath, dest string
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Download coin-executor binary from manifest URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}
			if dest == "" {
				dest = filepath.Join(".coin", "coin-executor")
			}
			if err := bootstrap.DownloadExecutor(m, dest); err != nil {
				return err
			}
			fmt.Printf("✓ executor downloaded: %s (v%s)\n", dest, m.Executor.Version)
			return nil
		},
	}
	cmd.Flags().StringVar(&manifestPath, "manifest", ".coin/manifest.json", "manifest path")
	cmd.Flags().StringVar(&dest, "dest", "", "destination path (default .coin/coin-executor)")
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print executor version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
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
