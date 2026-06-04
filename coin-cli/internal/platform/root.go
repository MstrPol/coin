package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirName           = "coin-platform"
	goldenPathsSubdir = "golden-paths"
	startersSubdir    = "starters"
	agentsSubdir      = "agents"
)

// Root — корень coin-platform (COIN_PLATFORM_DIR или поиск coin-platform/ вверх от cwd).
func Root() (string, error) {
	if dir := os.Getenv("COIN_PLATFORM_DIR"); dir != "" {
		return validateRoot(dir)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, dirName)
		if ok, _ := isPlatformRoot(candidate); ok {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("%s/ not found (set COIN_PLATFORM_DIR)", dirName)
}

func validateRoot(dir string) (string, error) {
	ok, reason := isPlatformRoot(dir)
	if !ok {
		return "", fmt.Errorf("COIN_PLATFORM_DIR=%s: %s", dir, reason)
	}
	return dir, nil
}

func isPlatformRoot(dir string) (bool, string) {
	if _, err := os.Stat(filepath.Join(dir, goldenPathsSubdir, "catalog.yaml")); err != nil {
		return false, "golden-paths/catalog.yaml not found"
	}
	if _, err := os.Stat(filepath.Join(dir, agentsSubdir, "catalog.yaml")); err != nil {
		return false, "agents/catalog.yaml not found"
	}
	return true, ""
}

// GoldenPathsDir — $COIN_PLATFORM_DIR/golden-paths.
func GoldenPathsDir() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, goldenPathsSubdir), nil
}

// StartersDir — $COIN_PLATFORM_DIR/starters.
func StartersDir() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, startersSubdir)
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("starters/ not found under %s", root)
	}
	return dir, nil
}

// AgentsCatalogPath — $COIN_PLATFORM_DIR/agents/catalog.yaml.
func AgentsCatalogPath() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, agentsSubdir, "catalog.yaml"), nil
}

// PlatformYAMLPath — $COIN_PLATFORM_DIR/platform.yaml.
func PlatformYAMLPath() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "platform.yaml"), nil
}
