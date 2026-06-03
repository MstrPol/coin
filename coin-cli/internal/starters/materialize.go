package starters

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/goldenpaths"
)

// Params — ответы визарда coin init.
type Params struct {
	Starter         string
	TemplateVersion string
	ProjectName     string
	GroupID         string
	Repository      string
	DockerCred        string
	QGMCred           string
	DestDir           string
	Force             bool
}

// Materialize копирует starter в dest и записывает .coin/config.yaml.
func Materialize(root string, p Params) error {
	if p.Starter == "" {
		return fmt.Errorf("starter is required")
	}
	if p.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}

	src := filepath.Join(root, p.Starter)
	if _, err := os.Stat(filepath.Join(src, ".coin", "config.yaml")); err != nil {
		return fmt.Errorf("starter %q not found in %s", p.Starter, root)
	}

	dest := p.DestDir
	if dest == "" {
		dest = "."
	}
	dest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	if err := copyTree(src, dest); err != nil {
		return err
	}

	version := p.TemplateVersion
	if version == "" {
		version, err = defaultTemplateVersion(p.Starter)
		if err != nil {
			return err
		}
	}

	return writeConfig(filepath.Join(dest, ".coin", "config.yaml"), p, version)
}

func defaultTemplateVersion(starter string) (string, error) {
	_, gpRoot, err := goldenpaths.Root()
	if err != nil {
		return "v1", nil
	}
	catalog, err := goldenpaths.LoadCatalog(os.DirFS(gpRoot))
	if err != nil {
		return "v1", nil
	}
	if e, ok := catalog.Paths[starter]; ok && e.Latest != "" {
		return e.Latest, nil
	}
	return "v1", nil
}

func copyTree(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("файл уже существует: %s (используйте --force)", target)
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func writeConfig(path string, p Params, version string) error {
	cfg := config.Config{
		Version: 1,
		Coin: config.CoinMeta{
			Template:        p.Starter,
			TemplateVersion: version,
		},
		Jenkins: config.JenkinsConfig{
			Credentials: config.Credentials{
				Docker: p.DockerCred,
				QGM:    p.QGMCred,
			},
		},
		Project: config.Project{
			Name:       p.ProjectName,
			GroupID:    p.GroupID,
			Repository: p.Repository,
		},
		RN: config.RNConfig{
			ServiceURL: "https://qgm.example.com",
		},
	}

	// container — из эталона starter (port/command зависят от стека)
	if existing, err := readContainerBlock(path); err == nil && existing.Port > 0 {
		cfg.Container = existing
	} else {
		cfg.Container = defaultContainer(p.Starter)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func readContainerBlock(path string) (config.Container, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return config.Container{}, err
	}
	var partial struct {
		Container config.Container `yaml:"container"`
	}
	if err := yaml.Unmarshal(raw, &partial); err != nil {
		return config.Container{}, err
	}
	return partial.Container, nil
}

func defaultContainer(starter string) config.Container {
	switch {
	case strings.HasPrefix(starter, "go"):
		return config.Container{Port: 8080, Command: []string{"/app/app"}}
	case strings.HasPrefix(starter, "java"):
		return config.Container{Port: 8080, Command: []string{"java", "-jar", "/app/app.jar"}}
	default:
		return config.Container{Port: 8080, Command: []string{"python", "-m", "my_service"}}
	}
}

// MaterializeForce как Materialize, но перезаписывает существующие файлы.
func MaterializeForce(root string, p Params) error {
	if p.Starter == "" || p.ProjectName == "" {
		return fmt.Errorf("starter and project name are required")
	}
	src := filepath.Join(root, p.Starter)
	dest, err := filepath.Abs(p.DestDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	if err := copyTreeForce(src, dest); err != nil {
		return err
	}
	version := p.TemplateVersion
	if version == "" {
		version, _ = defaultTemplateVersion(p.Starter)
	}
	return writeConfig(filepath.Join(dest, ".coin", "config.yaml"), p, version)
}

func copyTreeForce(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil || rel == "." {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}

// DestLooksInitialized — в каталоге уже есть Coin config.
func DestLooksInitialized(dest string) bool {
	_, err := os.Stat(filepath.Join(dest, ".coin", "config.yaml"))
	return err == nil
}
