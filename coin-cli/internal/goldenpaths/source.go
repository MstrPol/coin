package goldenpaths

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"coin.local/coin-cli/internal/platform"
)

const catalogFile = "catalog.yaml"

// Root возвращает fs.FS корня golden-paths.
//
// Источники (COIN_GOLDEN_PATHS_SOURCE):
//   - local (default): COIN_PLATFORM_DIR/golden-paths, COIN_GOLDEN_PATHS_DIR или поиск вверх от cwd
//   - nexus: COIN_GOLDEN_PATHS_URL — tarball (см. Fetch)
func Root() (fs.FS, string, error) {
	source := strings.ToLower(os.Getenv("COIN_GOLDEN_PATHS_SOURCE"))
	if source == "" {
		source = "local"
	}

	switch source {
	case "local":
		dir, err := localDir()
		if err != nil {
			return nil, "", err
		}
		return os.DirFS(dir), dir, nil
	case "nexus", "url":
		dir, err := FetchFromURL(os.Getenv("COIN_GOLDEN_PATHS_URL"))
		if err != nil {
			return nil, "", err
		}
		return os.DirFS(dir), dir, nil
	default:
		return nil, "", fmt.Errorf("unknown COIN_GOLDEN_PATHS_SOURCE %q (use local or nexus)", source)
	}
}

func localDir() (string, error) {
	if dir := os.Getenv("COIN_GOLDEN_PATHS_DIR"); dir != "" {
		if _, err := os.Stat(filepath.Join(dir, catalogFile)); err == nil {
			return dir, nil
		}
		return "", fmt.Errorf("COIN_GOLDEN_PATHS_DIR=%s: catalog.yaml not found", dir)
	}

	if dir, err := platform.GoldenPathsDir(); err == nil {
		return dir, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		for _, candidate := range []string{
			filepath.Join(dir, "coin-platform", "golden-paths"),
			filepath.Join(dir, "golden-paths"),
		} {
			if _, err := os.Stat(filepath.Join(candidate, catalogFile)); err == nil {
				return candidate, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("golden-paths/ not found (set COIN_PLATFORM_DIR or COIN_GOLDEN_PATHS_DIR)")
}

// LoadCatalog читает catalog.yaml из root FS.
func LoadCatalog(root fs.FS) (*Catalog, error) {
	data, err := fs.ReadFile(root, catalogFile)
	if err != nil {
		return nil, fmt.Errorf("read catalog: %w", err)
	}
	var paths map[string]CatalogEntry
	if err := yaml.Unmarshal(data, &paths); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}
	return &Catalog{Paths: paths}, nil
}

// SourceLabel — человекочитаемое описание источника каталога (для coin validate).
func SourceLabel() string {
	source := strings.ToLower(os.Getenv("COIN_GOLDEN_PATHS_SOURCE"))
	if source == "" {
		source = "local"
	}
	switch source {
	case "local":
		if dir := os.Getenv("COIN_GOLDEN_PATHS_DIR"); dir != "" {
			return fmt.Sprintf("local (%s)", dir)
		}
		if dir, err := platform.GoldenPathsDir(); err == nil {
			return fmt.Sprintf("platform (%s)", dir)
		}
		return "local"
	case "nexus", "url":
		if url := os.Getenv("COIN_GOLDEN_PATHS_URL"); url != "" {
			return fmt.Sprintf("%s (%s)", source, url)
		}
		return source
	default:
		return source
	}
}
