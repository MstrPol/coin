package scanner

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type coinConfig struct {
	Coin struct {
		GoldenPath      string `yaml:"goldenPath"`
		Version         string `yaml:"version"`
		Template        string `yaml:"template"`
		TemplateVersion string `yaml:"templateVersion"`
	} `yaml:"coin"`
	Project struct {
		Name string `yaml:"name"`
	} `yaml:"project"`
}

type ParsedConfig struct {
	Project    string
	GoldenPath string
	Version    string
}

func ParseConfig(raw []byte, repoName string) (ParsedConfig, error) {
	var cfg coinConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return ParsedConfig{}, ErrBadConfig
	}
	if cfg.Coin.Template != "" || cfg.Coin.TemplateVersion != "" {
		return ParsedConfig{}, ErrConfigV1
	}
	gp := strings.TrimSpace(cfg.Coin.GoldenPath)
	ver := strings.TrimSpace(cfg.Coin.Version)
	if gp == "" || ver == "" {
		return ParsedConfig{}, ErrBadConfig
	}
	project := strings.TrimSpace(cfg.Project.Name)
	if project == "" {
		project = repoName
	}
	return ParsedConfig{
		Project:    project,
		GoldenPath: gp,
		Version:    ver,
	}, nil
}
