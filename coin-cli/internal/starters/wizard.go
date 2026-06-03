package starters

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
)

// RunWizard — интерактивный TUI-визард coin init (↑↓ для выбора, Enter для подтверждения).
func RunWizard(root string) (Params, error) {
	opts, err := WizardOptions(root)
	if err != nil {
		return Params{}, err
	}

	var (
		starterID        string
		projectName      = "my-service"
		groupID          = "com.example.team"
		repositoryChoice = "Nexus_PROD"
		customRepository string
		dockerCred       = "nexus-docker"
		qgmCred          = "qgm-svc-account"
		destDir          string
	)

	starterOptions := make([]huh.Option[string], len(opts))
	for i, o := range opts {
		label := o.Title
		if o.Description != "" {
			label = o.Title + " · " + o.Description
		}
		starterOptions[i] = huh.NewOption(label, o.ID)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Golden path").
				Description("Профиль доставки — toolchain, сборка, публикация").
				Options(starterOptions...).
				Value(&starterID),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Имя сервиса").
				Description("project.name — artifactId, имя образа").
				Placeholder("my-service").
				Value(&projectName).
				Validate(required("имя сервиса")),
			huh.NewInput().
				Title("Group ID").
				Description("project.groupId — домен команды в реестре").
				Placeholder("com.example.team").
				Value(&groupID).
				Validate(required("group ID")),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Репозиторий Nexus").
				Description("project.repository — логическое имя в организации").
				Options(
					huh.NewOption("Nexus_PROD", "Nexus_PROD"),
					huh.NewOption("Nexus_SNAPSHOT", "Nexus_SNAPSHOT"),
					huh.NewOption("Другой…", "custom"),
				).
				Value(&repositoryChoice),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Jenkins credential — Docker registry").
				Description("jenkins.credentials.docker").
				Placeholder("nexus-docker").
				Value(&dockerCred).
				Validate(required("docker credential")),
			huh.NewInput().
				Title("Jenkins credential — QGM").
				Description("jenkins.credentials.qgm (можно оставить пустым)").
				Placeholder("qgm-svc-account").
				Value(&qgmCred),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Каталог проекта").
				Description("Пусто → ./<имя сервиса>").
				Placeholder("./my-service").
				Value(&destDir),
		),
	).WithTheme(huh.ThemeBase16())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return Params{}, fmt.Errorf("отменено")
		}
		return Params{}, err
	}

	projectName = strings.TrimSpace(projectName)
	if destDir = strings.TrimSpace(destDir); destDir == "" {
		destDir = "./" + projectName
	}

	repository := repositoryChoice
	if repositoryChoice == "custom" {
		if err := huh.NewInput().
			Title("Имя репозитория Nexus").
			Placeholder("Nexus_PROD").
			Value(&customRepository).
			Validate(required("имя репозитория")).
			Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return Params{}, fmt.Errorf("отменено")
			}
			return Params{}, err
		}
		repository = strings.TrimSpace(customRepository)
	}

	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return Params{}, err
	}

	if DestLooksInitialized(destAbs) {
		return Params{}, fmt.Errorf("в %s уже есть .coin/config.yaml", destAbs)
	}

	force := false
	if entries, _ := os.ReadDir(destAbs); len(entries) > 0 {
		var overwrite bool
		if err := huh.NewConfirm().
			Title("Каталог не пустой").
			Description(fmt.Sprintf("%s — добавить/перезаписать файлы стартера?", destAbs)).
			Value(&overwrite).
			Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return Params{}, fmt.Errorf("отменено")
			}
			return Params{}, err
		}
		if !overwrite {
			return Params{}, fmt.Errorf("отменено")
		}
		force = true
	}

	var create bool
	summary := fmt.Sprintf(
		"golden path: %s\nproject: %s\nкаталог: %s",
		starterID, projectName, destAbs,
	)
	if err := huh.NewConfirm().
		Title("Создать проект?").
		Description(summary).
		Affirmative("Да").
		Negative("Нет").
		Value(&create).
		Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return Params{}, fmt.Errorf("отменено")
		}
		return Params{}, err
	}
	if !create {
		return Params{}, fmt.Errorf("отменено")
	}

	return Params{
		Starter:     starterID,
		ProjectName: projectName,
		GroupID:     strings.TrimSpace(groupID),
		Repository:  repository,
		DockerCred:  strings.TrimSpace(dockerCred),
		QGMCred:     strings.TrimSpace(qgmCred),
		DestDir:     destAbs,
		Force:       force,
	}, nil
}

func required(label string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s обязательно", label)
		}
		return nil
	}
}

// IsInteractive — stdin подключён к терминалу.
func IsInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
