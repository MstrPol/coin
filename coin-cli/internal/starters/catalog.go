package starters

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"coin.local/coin-cli/internal/platform"
)

// Root возвращает путь к каталогу starters.
func Root() (string, error) {
	if dir := os.Getenv("COIN_STARTERS_DIR"); dir != "" {
		if isStarterRoot(dir) {
			return dir, nil
		}
		return "", fmt.Errorf("COIN_STARTERS_DIR=%s: каталог стартеров не найден", dir)
	}

	if dir, err := platform.StartersDir(); err == nil {
		return dir, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		for _, candidate := range []string{
			filepath.Join(dir, "coin-platform", "starters"),
			filepath.Join(dir, "coin-starters"),
			filepath.Join(dir, "starters"),
		} {
			if isStarterRoot(candidate) {
				return candidate, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("starters/ not found (set COIN_PLATFORM_DIR or COIN_STARTERS_DIR)")
}

func isStarterRoot(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, e.Name(), ".coin", "config.yaml")); err == nil {
			return true
		}
	}
	return false
}

// List возвращает имена доступных стартеров (go-app, python-uv-app, …).
func List(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, name, ".coin", "config.yaml")); err != nil {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, fmt.Errorf("в %s нет стартеров", root)
	}
	return names, nil
}

// Option — пункт выбора в визарде coin init.
type Option struct {
	ID          string
	Title       string
	Description string
}

// WizardOptions возвращает стартеры с подписями для интерактивного меню.
func WizardOptions(root string) ([]Option, error) {
	names, err := List(root)
	if err != nil {
		return nil, err
	}
	opts := make([]Option, 0, len(names))
	for _, name := range names {
		opts = append(opts, Option{
			ID:          name,
			Title:       wizardTitle(name),
			Description: wizardDescription(name),
		})
	}
	return opts, nil
}

func wizardTitle(name string) string {
	switch name {
	case "go-app":
		return "Go — приложение (OCI image)"
	case "java-gradle-app":
		return "Java 17 + Gradle — приложение"
	case "java-maven-app":
		return "Java 17 + Maven — приложение"
	case "python-uv-app":
		return "Python + uv — приложение"
	case "python-pip-app":
		return "Python + pip — приложение"
	default:
		return name
	}
}

func wizardDescription(name string) string {
	switch name {
	case "go-app":
		return "go test, container build, Docker registry"
	case "java-gradle-app", "java-maven-app":
		return "JUnit, JAR + OCI image, Docker registry"
	case "python-uv-app":
		return "uv + pytest, OCI image, Docker registry"
	case "python-pip-app":
		return "pip + pytest, OCI image, Docker registry"
	default:
		return "golden path: " + name
	}
}
