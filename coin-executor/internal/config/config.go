package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultPath = ".coin/config.yaml"

type Config struct {
	Coin    CoinMeta      `yaml:"coin"`
	Jenkins JenkinsConfig `yaml:"jenkins"`
	Project Project       `yaml:"project"`
}

type CoinMeta struct {
	GoldenPath string `yaml:"goldenPath"`
	Version    string `yaml:"version"`
}

type JenkinsConfig struct {
	Credentials Credentials `yaml:"credentials"`
}

type Credentials struct {
	Docker string `yaml:"docker"`
}

type Project struct {
	Name       string `yaml:"name"`
	GroupID    string `yaml:"groupId"`
	Repository string `yaml:"repository"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Coin.GoldenPath == "" {
		return fmt.Errorf("coin.goldenPath is required (config v2)")
	}
	if cfg.Coin.Version == "" {
		return fmt.Errorf("coin.version is required (config v2)")
	}
	if cfg.Project.Name == "" {
		return fmt.Errorf("project.name is required")
	}
	if cfg.Jenkins.Credentials.Docker == "" {
		return fmt.Errorf("jenkins.credentials.docker is required")
	}
	return nil
}
