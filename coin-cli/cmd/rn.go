package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/releasenotes"
	"coin.local/coin-cli/internal/versioning"
)

// rnCmd — корневая группа команд «coin rn».
var rnCmd = &cobra.Command{
	Use:   "rn",
	Short: "Операции с release notes",
	Long: `Команды для работы с release notes (интеграция с QGM API).

Примеры:
  coin rn generate            # сгенерировать и сохранить
  coin rn generate --dry-run  # показать сводку без сохранения`,
}

// rnGenerateCmd — coin rn generate.
var rnGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Сгенерировать release notes из git-истории",
	Long: `Собирает Jira-тикеты из smart-коммитов и сохраняет JSON
в .coin/temp/release-notes.json.

Диапазон коммитов определяется АВТОМАТИЧЕСКИ по модели ветвления:
  - Текущая версия (coin version) → парсится base (major.minor.patch).
  - Нижняя граница — последний RC-тег с другим base (предыдущий релиз).
    Если предыдущего релиза нет (новый проект) — с самого первого коммита.
  - Верхняя граница — HEAD.

Если текущая версия rc-5, в release notes попадут ВСЕ тикеты
данного релиза (rc-1 … rc-5).`,
	RunE: runRNGenerate,
}

func init() {
	flags := rnGenerateCmd.Flags()
	flags.String("release-link", "", "ссылка на Jira-задачу «Release 2.0»")
	flags.String("output", ".coin/temp/release-notes.json", "путь для сохранения JSON")
	flags.String("config", config.DefaultPath, "путь до .coin/config.yaml")
	flags.Bool("dry-run", false, "вывести сводку без сохранения на диск")

	rnCmd.AddCommand(rnGenerateCmd)
}

func runRNGenerate(cmd *cobra.Command, _ []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	releaseLink, _ := cmd.Flags().GetString("release-link")
	output, _ := cmd.Flags().GetString("output")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if cfg.Project.Repository == "" {
		return fmt.Errorf("project.repository не задан в %s", cfgPath)
	}
	if cfg.Project.GroupID == "" {
		return fmt.Errorf("project.groupId не задан в %s", cfgPath)
	}

	version, err := versioning.CurrentVersion()
	if err != nil {
		return fmt.Errorf("версия: %w", err)
	}

	from, fromLabel, err := detectFromTag(version)
	if err != nil {
		return fmt.Errorf("авто-определение диапазона: %w", err)
	}

	opts := releasenotes.Options{
		Repository:     cfg.Project.Repository,
		GroupID:        cfg.Project.GroupID,
		ArtifactID:     cfg.Project.Name,
		Version:        version,
		From:           from, // "" → первый релиз, иначе тег предыдущего релиза
		CodeRepository: cfg.RN.CodeRepository,
		ReleaseLink:    releaseLink,
	}

	payload, err := releasenotes.Generate(opts)
	if err != nil {
		return fmt.Errorf("генерация: %w", err)
	}

	if dryRun {
		fmt.Printf("artifact  : %s/%s/%s @ %s\n",
			payload.Repository, payload.GroupID, payload.ArtifactID, payload.Version)
		fmt.Printf("диапазон  : %s\n", fromLabel)
		fmt.Printf("тикетов   : %d\n", len(payload.ReleaseNotes))
		for _, rn := range payload.ReleaseNotes {
			fmt.Printf("  • %s — %s\n", rn.Issue, rn.Summary)
		}
		fmt.Printf("участников: %d\n", countContributors(payload))
		fmt.Println("(dry-run: файл не сохранён)")
		return nil
	}

	if err := releasenotes.SaveToFile(payload, output); err != nil {
		return fmt.Errorf("сохранение: %w", err)
	}

	fmt.Printf("Release notes сохранены: %s\n", output)
	fmt.Printf("  Версия     : %s\n", payload.Version)
	fmt.Printf("  Диапазон   : %s\n", fromLabel)
	fmt.Printf("  Тикетов    : %d\n", len(payload.ReleaseNotes))
	fmt.Printf("  Участников : %d\n", countContributors(payload))

	return nil
}

// detectFromTag определяет нижнюю границу диапазона коммитов по версии.
//
//  1. Из версии (напр. "1.5.0-PROJ-404-rc-5") берём base="1.5.0".
//  2. Ищем последний RC-тег с другим base — это предыдущий выпущенный релиз.
//  3. Если такого нет (новый проект / первый релиз) — вся история.
func detectFromTag(version string) (tag, label string, err error) {
	base := versionBase(version)
	if base == "" {
		return "", "(начало репозитория — новый проект)", nil
	}

	prevTag, err := versioning.LatestReleaseTag(base)
	if err != nil {
		return "", "", err
	}
	if prevTag == "" {
		return "", "(начало репозитория — первый релиз)", nil
	}

	return prevTag, prevTag + "..HEAD", nil
}

// versionBase извлекает major.minor.patch из строки версии.
// "1.5.0-PROJ-404-rc-5" → "1.5.0"
func versionBase(version string) string {
	parts := strings.SplitN(version, "-", 2)
	if len(parts) >= 1 && isSemver(parts[0]) {
		return parts[0]
	}
	return ""
}

// isSemver проверяет формат major.minor.patch.
func isSemver(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if len(p) == 0 {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func countContributors(p *releasenotes.Payload) int {
	seen := map[string]bool{}
	for _, list := range p.Contributors {
		for _, c := range list {
			seen[c.Email] = true
		}
	}
	return len(seen)
}
