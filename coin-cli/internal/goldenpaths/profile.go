package goldenpaths

import "fmt"

// Profile — platform-owned сценарий golden path (coin-golden-paths/<name>/vN/profile.yaml).
type Profile struct {
	Agent struct {
		Stack   string            `yaml:"stack"`
		Runtime map[string]string `yaml:"runtime"`
		Rev     int               `yaml:"rev"` // pin → agents/catalog.yaml (image tag …-rN)
	} `yaml:"agent"`
	CoinCli struct {
		Version string `yaml:"version"` // pin → Nexus Maven coin-cli zip
	} `yaml:"coinCli"`
	Build struct {
		Type       string `yaml:"type"`       // container | package
		Dockerfile string `yaml:"dockerfile"` // имя файла в bundle, default Dockerfile
	} `yaml:"build"`
	Publish struct {
		Kind string `yaml:"kind"` // registry | maven | pypi
		When string `yaml:"when"` // tag | branch | always | never
	} `yaml:"publish"`
	Pipeline struct {
		Test    StageProfile `yaml:"test"`
		Build   StageProfile `yaml:"build"`
		Publish StageProfile `yaml:"publish"`
	} `yaml:"pipeline"`
}

// StageProfile — дефолты стадий из profile (перекрываются project.pipeline).
type StageProfile struct {
	Enabled *bool `yaml:"enabled"`
}

// CatalogEntry — policy golden path из catalog.yaml.
type CatalogEntry struct {
	Stack      string   `yaml:"stack"`
	Latest     string   `yaml:"latest"`
	Minimum    string   `yaml:"minimum"`
	Deprecated []string `yaml:"deprecated,omitempty"`
}

// Catalog — разобранный catalog.yaml.
type Catalog struct {
	Paths map[string]CatalogEntry
}

// DeprecationWarning возвращает текст предупреждения, если версия профиля deprecated.
func (c *Catalog) DeprecationWarning(name, version string) string {
	if c == nil {
		return ""
	}
	entry, ok := c.Paths[name]
	if !ok || !entry.isDeprecated(version) {
		return ""
	}
	if entry.Latest != "" && entry.Latest != version {
		return fmt.Sprintf("версия %s снята с поддержки — перейдите на %s", version, entry.Latest)
	}
	return fmt.Sprintf("версия %s снята с поддержки", version)
}

func (e CatalogEntry) isDeprecated(version string) bool {
	for _, v := range e.Deprecated {
		if v == version {
			return true
		}
	}
	return false
}
