package dockerfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/goldenpaths"
)

const GeneratedDir = ".coin/generated"
const GeneratedPath = ".coin/generated/Dockerfile"

func Render(cfg *config.Config, bundle *goldenpaths.Bundle) (string, error) {
	if _, err := os.Stat("Dockerfile"); err == nil {
		return "", fmt.Errorf(
			"Dockerfile найден в корне репозитория. " +
				"При container-сборке Dockerfile управляется Coin. Удалите Dockerfile из репозитория.",
		)
	}

	if bundle.BuildType() != "container" {
		return "", fmt.Errorf("template %s/%s build.type=%q, container Dockerfile не требуется",
			bundle.Name, bundle.Version, bundle.BuildType())
	}

	tmpl, err := bundle.Dockerfile()
	if err != nil {
		return "", fmt.Errorf("dockerfile template: %w", err)
	}

	rendered := render(tmpl, cfg, bundle)

	if err := os.MkdirAll(GeneratedDir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(GeneratedPath, []byte(rendered), 0o644); err != nil {
		return "", err
	}

	absPath, _ := filepath.Abs(GeneratedPath)
	return absPath, nil
}

func render(tmpl string, cfg *config.Config, bundle *goldenpaths.Bundle) string {
	pythonVersion := bundle.RuntimeVersion("python", cfg.RuntimeVersion("python", "3.13"))
	javaVersion := bundle.RuntimeVersion("java", cfg.RuntimeVersion("java", "21"))
	goVersion := bundle.RuntimeVersion("go", cfg.RuntimeVersion("go", "1.22"))

	replacer := strings.NewReplacer(
		"{{PYTHON_VERSION}}", pythonVersion,
		"{{JAVA_VERSION}}", javaVersion,
		"{{GO_VERSION}}", goVersion,
		"{{APP_PORT}}", bundle.ContainerPort(),
		"{{APP_CMD}}", bundle.ContainerCommand(),
	)
	return replacer.Replace(tmpl)
}
