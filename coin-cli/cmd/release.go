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
	Short: "Bump version tag on HEAD of current branch (patch|minor|major|rc)",
	Long: `Создать и запушить следующий тег версии.

  patch  — v1.4.1 → v1.4.2   (баг-фикс, ПСИ-фикс после финального релиза)
  minor  — v1.4.x → v1.5.0   (новая фича)
  major  — v1.x.x → v2.0.0   (breaking change)
  rc     — v1.5.0-rc.1 → v1.5.0-rc.2   (итерация ПСИ; базовая версия берётся из --base)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		bumpType, _ := cmd.Flags().GetString("type")
		if bumpType != "patch" && bumpType != "minor" && bumpType != "major" && bumpType != "rc" {
			return fmt.Errorf("--type должен быть patch, minor, major или rc")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		prefix := cfg.TagPrefix()

		// ── RC: отдельная ветка логики ───────────────────────────────────────
		if bumpType == "rc" {
			return bumpRC(cmd, prefix, dryRun)
		}

		// ── patch / minor / major ─────────────────────────────────────────────
		latest, err := versioning.LatestReleaseTag(prefix, "")
		if err != nil {
			return err
		}

		var nextVersion string
		if latest == "" {
			switch bumpType {
			case "major":
				nextVersion = "1.0.0"
			case "minor":
				nextVersion = "0.1.0"
			default:
				nextVersion = "0.0.1"
			}
		} else {
			// Для bump берём только финальные теги (без -rc.N)
			baseVersion := strings.TrimPrefix(latest, prefix)
			// Если последний тег — RC, используем его базу
			if rcBase := rcBaseVersion(baseVersion); rcBase != "" {
				baseVersion = rcBase
			}
			nextVersion, err = bump(baseVersion, bumpType)
			if err != nil {
				return err
			}
		}

		nextTag := prefix + nextVersion
		fmt.Printf("Предыдущий тег: %s\n", orNone(latest))
		fmt.Printf("Следующий тег:  %s\n", nextTag)

		if dryRun {
			fmt.Println("[dry-run] тег не создан")
			return nil
		}

		if err := gitTag(nextTag); err != nil {
			return fmt.Errorf("не удалось создать тег %s: %w", nextTag, err)
		}
		fmt.Printf("✓ тег %s создан и запушен\n", nextTag)
		return nil
	},
}

// bumpRC создаёт следующий release candidate для заданной базовой версии.
// Базовая версия берётся из --base (например 1.5.0), или вычисляется из последнего тега.
func bumpRC(cmd *cobra.Command, prefix string, dryRun bool) error {
	base, _ := cmd.Flags().GetString("base")

	if base == "" {
		// Вычислить базу: bump minor от последнего финального тега
		latest, err := versioning.LatestReleaseTag(prefix, "")
		if err != nil {
			return err
		}
		if latest == "" {
			base = "0.1.0"
		} else {
			raw := strings.TrimPrefix(latest, prefix)
			if b := rcBaseVersion(raw); b != "" {
				base = b
			} else {
				base, err = bump(raw, "minor")
				if err != nil {
					return err
				}
			}
		}
		fmt.Printf("База RC (вычислена): %s\n", base)
	}

	n, err := versioning.NextRCNumber(prefix, base)
	if err != nil {
		return err
	}

	nextTag := fmt.Sprintf("%s%s-rc.%d", prefix, base, n)
	fmt.Printf("Следующий RC тег: %s\n", nextTag)

	if dryRun {
		fmt.Println("[dry-run] тег не создан")
		return nil
	}

	if err := gitTag(nextTag); err != nil {
		return fmt.Errorf("не удалось создать тег %s: %w", nextTag, err)
	}
	fmt.Printf("✓ тег %s создан и запушен\n", nextTag)
	return nil
}

func init() {
	releaseBumpCmd.Flags().String("type", "", "bump type: patch, minor, major или rc (required)")
	_ = releaseBumpCmd.MarkFlagRequired("type")
	releaseBumpCmd.Flags().Bool("dry-run", false, "показать следующий тег без создания")
	releaseBumpCmd.Flags().String("base", "", "базовая версия для RC (например 1.5.0); вычисляется автоматически если не задана")
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

// rcBaseVersion возвращает базовую версию из RC-строки (например "1.5.0-rc.2" → "1.5.0").
// Возвращает пустую строку если строка не является RC.
var rcRe = regexp.MustCompile(`^(\d+\.\d+\.\d+)-rc\.\d+$`)

func rcBaseVersion(version string) string {
	m := rcRe.FindStringSubmatch(version)
	if m == nil {
		return ""
	}
	return m[1]
}

func orNone(s string) string {
	if s == "" {
		return "(нет тегов)"
	}
	return s
}
