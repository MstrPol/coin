package dockerfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"coin.local/coin-cli/embed"
	"coin.local/coin-cli/internal/config"
)

const GeneratedDir = ".coin/generated"
const GeneratedPath = ".coin/generated/Dockerfile"

func Render(cfg *config.Config) (string, error) {
	// Запрещаем Dockerfile в корне репозитория
	if _, err := os.Stat("Dockerfile"); err == nil {
		return "", fmt.Errorf(
			"Dockerfile найден в корне репозитория. " +
				"При target: container Dockerfile управляется Coin централизованно. " +
				"Удалите Dockerfile из репозитория сервиса.",
		)
	}

	templateName := cfg.Pipeline.Build.DockerfileTemplate
	if templateName == "" {
		templateName = cfg.Project.Stack
	}

	tmpl, err := embed.DockerfileTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("dockerfile template %q not found: %w", templateName, err)
	}

	rendered := render(tmpl, cfg)

	if err := os.MkdirAll(GeneratedDir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(GeneratedPath, []byte(rendered), 0644); err != nil {
		return "", err
	}

	absPath, _ := filepath.Abs(GeneratedPath)
	return absPath, nil
}

func render(tmpl string, cfg *config.Config) string {
	pythonVersion := cfg.Runtime["python"]
	if pythonVersion == "" {
		pythonVersion = "3.13"
	}
	javaVersion := cfg.Runtime["java"]
	if javaVersion == "" {
		javaVersion = "21"
	}
	goVersion := cfg.Runtime["go"]
	if goVersion == "" {
		goVersion = "1.22"
	}

	port := "8080"
	if cfg.Container.Port > 0 {
		port = strconv.Itoa(cfg.Container.Port)
	}

	cmd := `["python", "-m", "app"]`
	if len(cfg.Container.Command) > 0 {
		parts := make([]string, len(cfg.Container.Command))
		for i, p := range cfg.Container.Command {
			parts[i] = `"` + p + `"`
		}
		cmd = "[" + strings.Join(parts, ", ") + "]"
	}

	replacer := strings.NewReplacer(
		"{{PYTHON_VERSION}}", pythonVersion,
		"{{JAVA_VERSION}}", javaVersion,
		"{{GO_VERSION}}", goVersion,
		"{{APP_PORT}}", port,
		"{{APP_CMD}}", cmd,
	)
	return replacer.Replace(tmpl)
}
