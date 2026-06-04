package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"coin.local/coin-cli/internal/starters"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Создать новый проект (интерактивный визард)",
	Long: `Скопирует скелетон из coin-starters/ и сформирует .coin/config.yaml.

Интерактивный режим: ↑↓ для выбора, Enter для подтверждения.

Пример (неинтерактивно):
  coin init --starter python-uv-app --name payments-api --dir ./payments-api`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().String("starters-dir", "", "путь к coin-starters/ (default: поиск вверх или COIN_STARTERS_DIR)")
	initCmd.Flags().String("starter", "", "golden path / имя стартера (go-app, python-uv-app, …)")
	initCmd.Flags().String("name", "", "project.name")
	initCmd.Flags().String("group-id", "", "project.groupId")
	initCmd.Flags().String("repository", "", "project.repository")
	initCmd.Flags().String("docker-cred", "", "jenkins.credentials.docker")
	initCmd.Flags().String("dir", "", "каталог назначения")
	initCmd.Flags().Bool("force", false, "перезаписать существующие файлы")
	initCmd.Flags().Bool("yes", false, "не задавать вопросов (требуются --starter и --name)")
}

func runInit(cmd *cobra.Command, _ []string) error {
	root, err := resolveStartersRoot(cmd)
	if err != nil {
		return err
	}

	var params starters.Params
	force, _ := cmd.Flags().GetBool("force")

	if useNonInteractive(cmd) {
		params, err = paramsFromFlags(cmd)
		if err != nil {
			return err
		}
	} else {
		if !starters.IsInteractive() {
			return fmt.Errorf("интерактивный режим недоступен (stdin не TTY); задайте --starter, --name и --yes")
		}
		params, err = starters.RunWizard(root)
		if err != nil {
			return err
		}
	}

	if params.DestDir == "" {
		params.DestDir = "./" + params.ProjectName
	}
	params.DestDir, err = filepath.Abs(params.DestDir)
	if err != nil {
		return err
	}

	if !force && starters.DestLooksInitialized(params.DestDir) {
		return fmt.Errorf("%s уже содержит .coin/config.yaml (используйте --force)", params.DestDir)
	}

	force = force || params.Force
	if force {
		err = starters.MaterializeForce(root, params)
	} else {
		err = starters.Materialize(root, params)
	}
	if err != nil {
		return err
	}

	fmt.Printf("\n✓ проект создан: %s\n", params.DestDir)
	fmt.Printf("  golden path: %s\n", params.Starter)
	fmt.Printf("  дальше: cd %s && coin validate\n", params.DestDir)
	return nil
}

func resolveStartersRoot(cmd *cobra.Command) (string, error) {
	if dir, _ := cmd.Flags().GetString("starters-dir"); dir != "" {
		if err := os.Setenv("COIN_STARTERS_DIR", dir); err != nil {
			return "", err
		}
	}
	return starters.Root()
}

func useNonInteractive(cmd *cobra.Command) bool {
	yes, _ := cmd.Flags().GetBool("yes")
	starter, _ := cmd.Flags().GetString("starter")
	name, _ := cmd.Flags().GetString("name")
	return yes || (starter != "" && name != "")
}

func paramsFromFlags(cmd *cobra.Command) (starters.Params, error) {
	starter, _ := cmd.Flags().GetString("starter")
	name, _ := cmd.Flags().GetString("name")
	if starter == "" || name == "" {
		return starters.Params{}, fmt.Errorf("--starter и --name обязательны в неинтерактивном режиме")
	}

	groupID, _ := cmd.Flags().GetString("group-id")
	if groupID == "" {
		groupID = "com.example.team"
	}
	repository, _ := cmd.Flags().GetString("repository")
	if repository == "" {
		repository = "Nexus_PROD"
	}
	dockerCred, _ := cmd.Flags().GetString("docker-cred")
	if dockerCred == "" {
		dockerCred = "nexus-docker"
	}
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir = "./" + name
	}
	return starters.Params{
		Starter:     starter,
		ProjectName: name,
		GroupID:     groupID,
		Repository:  repository,
		DockerCred:  dockerCred,
		DestDir:     dir,
	}, nil
}
