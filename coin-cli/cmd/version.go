package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/versioning"
)

// versionCmd — coin version
// Выводит текущую версию (последний тег или 0.0.1 если тегов нет).
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current COIN_VERSION",
	Long: `Show the current version derived from git tags.

If no tags exist yet (new project) returns 0.0.1.
HEAD is tagged → that tag's version.
Otherwise → latest tag's version.

Use 'coin version bump' to create the next version tag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := versioning.CurrentVersion()
		if err != nil {
			return err
		}
		fmt.Println(v)
		return nil
	},
}

// versionBumpCmd — coin version bump [major|minor|patch]
var versionBumpCmd = &cobra.Command{
	Use:   "bump [major|minor|patch]",
	Short: "Create the next version tag",
	Long: `Compute the next version and create a git tag.

Behaviour:
  • If a series for the current branch already exists (same Jira ID + type),
    the base version is kept and only N is incremented.
    Example: v1.5.0-PROJ-404-rc-1 exists → creates v1.5.0-PROJ-404-rc-2.

  • If no series exists yet, applies the bump level to the latest base version
    and starts N at 1.
    Example: latest base 1.5.0, bump patch → v1.5.1-PROJ-101-snapshot-1.

  --type rc is only allowed on release/* branches.
  --type snapshot is the default and works on any branch.`,
	Args: cobra.ExactArgs(1),
	ValidArgs: []string{"major", "minor", "patch"},
	RunE: func(cmd *cobra.Command, args []string) error {
		bumpLevel := args[0]
		if bumpLevel != "major" && bumpLevel != "minor" && bumpLevel != "patch" {
			return fmt.Errorf("уровень bump должен быть major, minor или patch; получено: %q", bumpLevel)
		}

		versionType, _ := cmd.Flags().GetString("type")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		branch := versioning.CurrentBranch()

		if versionType == "rc" && !strings.HasPrefix(branch, "release/") {
			return fmt.Errorf(
				"--type rc допускается только на ветке release/*\n"+
					"текущая ветка: %q\n"+
					"Создайте release/* ветку или используйте --type snapshot",
				branch,
			)
		}

		jiraID := versioning.BranchJiraID(branch)

		nextTag, err := versioning.NextVersionTag(jiraID, bumpLevel, versionType)
		if err != nil {
			return err
		}

		fmt.Printf("Следующий тег: %s\n", nextTag)

		if dryRun {
			fmt.Println("[dry-run] тег не создан")
			return nil
		}

		if err := gitTag(nextTag); err != nil {
			return fmt.Errorf("не удалось создать тег %s: %w", nextTag, err)
		}
		fmt.Printf("✓ %s создан и запушен\n", nextTag)
		return nil
	},
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

func init() {
	versionBumpCmd.Flags().String("type", "snapshot", "тип версии: snapshot (по умолчанию) или rc")
	versionBumpCmd.Flags().Bool("dry-run", false, "показать тег без создания")
	versionBumpCmd.Flags().String("config", config.DefaultPath, "путь до .coin/config.yaml")

	versionCmd.AddCommand(versionBumpCmd)
}
