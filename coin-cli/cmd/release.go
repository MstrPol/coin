package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/versioning"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release operations",
}

var releaseBumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Bump version tag on HEAD of current branch (patch|minor|major)",
	RunE: func(cmd *cobra.Command, args []string) error {
		bumpType, _ := cmd.Flags().GetString("type")
		if bumpType != "patch" && bumpType != "minor" && bumpType != "major" {
			return fmt.Errorf("--type must be patch, minor or major")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		prefix := cfg.TagPrefix()

		// Найти последний релизный тег
		latest, err := versioning.LatestReleaseTag(prefix, "")
		if err != nil {
			return err
		}

		var nextVersion string
		if latest == "" {
			// Первый релиз
			switch bumpType {
			case "major":
				nextVersion = "1.0.0"
			case "minor":
				nextVersion = "0.1.0"
			default:
				nextVersion = "0.0.1"
			}
		} else {
			nextVersion, err = bump(strings.TrimPrefix(latest, prefix), bumpType)
			if err != nil {
				return err
			}
		}

		nextTag := prefix + nextVersion
		fmt.Printf("Предыдущий тег: %s\n", latest)
		fmt.Printf("Следующий тег:  %s\n", nextTag)

		if dryRun {
			fmt.Println("[dry-run] тег не создан")
			return nil
		}

		// Создать и запушить тег
		if err := gitTag(nextTag); err != nil {
			return fmt.Errorf("не удалось создать тег %s: %w", nextTag, err)
		}
		fmt.Printf("✓ тег %s создан и запушен\n", nextTag)
		return nil
	},
}

func init() {
	releaseBumpCmd.Flags().String("type", "", "bump type: patch, minor or major (required)")
	_ = releaseBumpCmd.MarkFlagRequired("type")
	releaseBumpCmd.Flags().Bool("dry-run", false, "показать следующий тег без создания")
	releaseBumpCmd.Flags().String("config", config.DefaultPath, "path to .coin/config.yaml")
	releaseCmd.AddCommand(releaseBumpCmd)
}

var semverRe = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

func bump(version, bumpType string) (string, error) {
	m := semverRe.FindStringSubmatch(version)
	if m == nil {
		return "", fmt.Errorf("не удалось распарсить semver %q", version)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])

	switch bumpType {
	case "major":
		major++
		minor, patch = 0, 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

func gitTag(tag string) error {
	run := func(args ...string) error {
		c := exec.Command("git", args...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}
	if err := run("tag", tag); err != nil {
		return err
	}
	return run("push", "origin", tag)
}
