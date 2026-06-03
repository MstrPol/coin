package goldenpaths

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// Bundle — разрешённый golden path: profile + файлы версии.
type Bundle struct {
	Name    string
	Version string
	Root    fs.FS
	RootDir string
	Profile Profile
	Catalog *Catalog
}

// Resolve загружает bundle по имени шаблона и версии (v1, v2, …).
func Resolve(name, version string) (*Bundle, error) {
	root, rootDir, err := Root()
	if err != nil {
		return nil, err
	}

	catalog, err := LoadCatalog(root)
	if err != nil {
		return nil, err
	}

	canonical := name
	if _, ok := catalog.Paths[canonical]; !ok {
		return nil, fmt.Errorf("unknown template %q (not in catalog)", name)
	}

	if version == "" {
		if e, ok := catalog.Paths[canonical]; ok && e.Latest != "" {
			version = e.Latest
		} else {
			version = "v1"
		}
	}

	base := path.Join(canonical, version)
	if _, err := fs.Stat(root, path.Join(base, "profile.yaml")); err != nil {
		return nil, fmt.Errorf("template %s/%s not found in catalog", canonical, version)
	}

	profile, err := loadProfile(root, base)
	if err != nil {
		return nil, err
	}

	if err := checkVersionPolicy(canonical, version, catalog.Paths); err != nil {
		return nil, err
	}

	return &Bundle{
		Name:    canonical,
		Version: version,
		Root:    root,
		RootDir: rootDir,
		Profile: *profile,
		Catalog: catalog,
	}, nil
}

func loadProfile(root fs.FS, base string) (*Profile, error) {
	data, err := fs.ReadFile(root, path.Join(base, "profile.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read profile.yaml: %w", err)
	}
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse profile.yaml: %w", err)
	}
	if p.Build.Dockerfile == "" {
		p.Build.Dockerfile = "Dockerfile"
	}
	if p.Build.Type == "" {
		p.Build.Type = "container"
	}
	if p.Publish.When == "" {
		p.Publish.When = "tag"
	}
	return &p, nil
}

func checkVersionPolicy(name, version string, catalog map[string]CatalogEntry) error {
	e, ok := catalog[name]
	if !ok || e.Minimum == "" {
		return nil
	}
	if compareVersion(version, e.Minimum) < 0 {
		return fmt.Errorf("template %s version %s below minimum %s", name, version, e.Minimum)
	}
	return nil
}

// compareVersion сравнивает v1, v2 (простой numeric suffix).
func compareVersion(a, b string) int {
	na := parseV(a)
	nb := parseV(b)
	if na == nb {
		return 0
	}
	if na < nb {
		return -1
	}
	return 1
}

func parseV(s string) int {
	s = strings.TrimPrefix(strings.ToLower(s), "v")
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// ReadFile читает файл из директории версии шаблона.
func (b *Bundle) ReadFile(rel string) ([]byte, error) {
	p := path.Join(b.Name, b.Version, rel)
	return fs.ReadFile(b.Root, p)
}

// Script возвращает скрипт стадии из bundle.
func (b *Bundle) Script(stage string) (string, error) {
	data, err := b.ReadFile(path.Join("scripts", stage+".sh"))
	if err != nil {
		return "", fmt.Errorf("script %s: %w", stage, err)
	}
	return string(data), nil
}

// Dockerfile возвращает содержимое managed Dockerfile.
func (b *Bundle) Dockerfile() (string, error) {
	name := b.Profile.Build.Dockerfile
	if name == "" {
		name = "Dockerfile"
	}
	data, err := b.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Stack возвращает toolchain stack из profile.
func (b *Bundle) Stack() string {
	return b.Profile.Agent.Stack
}

// BuildType container | package.
func (b *Bundle) BuildType() string {
	return b.Profile.Build.Type
}

// RuntimeVersion из profile с fallback.
func (b *Bundle) RuntimeVersion(key, fallback string) string {
	if v := b.Profile.Agent.Runtime[key]; v != "" {
		return v
	}
	return fallback
}

// StageEnabled — дефолт из profile (nil = true).
func (b *Bundle) StageEnabled(stage string, projectEnabled *bool) bool {
	if projectEnabled != nil {
		return *projectEnabled
	}
	var p *bool
	switch stage {
	case "test":
		p = b.Profile.Pipeline.Test.Enabled
	case "build":
		p = b.Profile.Pipeline.Build.Enabled
	case "publish":
		p = b.Profile.Pipeline.Publish.Enabled
	}
	if p == nil {
		return true
	}
	return *p
}
