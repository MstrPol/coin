package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultPath = ".coin/config.yaml"

// Config — контракт .coin/config.yaml проекта.
//
// Поведение сборки задаётся golden path (coin.template + templateVersion).
// Секция jenkins: — только для coin-lib (агент + credentials).
type Config struct {
	Version   int           `yaml:"version"`
	Coin      CoinMeta      `yaml:"coin"`
	Jenkins   JenkinsConfig `yaml:"jenkins"`
	Project   Project       `yaml:"project"`
	Container Container     `yaml:"container"`
	Pipeline  Pipeline      `yaml:"pipeline"` // optional overrides
	RN        RNConfig      `yaml:"rn"`
}

// JenkinsConfig — секция для Jenkins (coin-lib): credentials и optional runtime override.
type JenkinsConfig struct {
	Stack       string            `yaml:"stack,omitempty"`
	Runtime     map[string]string `yaml:"runtime,omitempty"`
	Credentials Credentials       `yaml:"credentials"`
}

// Credentials — Jenkins Credential IDs.
type Credentials struct {
	Docker string `yaml:"docker"`
	QGM    string `yaml:"qgm"`
	Nexus  string `yaml:"nexus"`
}

type CoinMeta struct {
	Template        string `yaml:"template"`
	TemplateVersion string `yaml:"templateVersion"`
}

type Project struct {
	Name       string `yaml:"name"`
	GroupID    string `yaml:"groupId"`
	Repository string `yaml:"repository"`
}

type RNConfig struct {
	ServiceURL     string `yaml:"serviceUrl"`
	CodeRepository string `yaml:"codeRepository"`
}

type Container struct {
	Port    int      `yaml:"port"`
	Command []string `yaml:"command"`
}

type Pipeline struct {
	Test    Stage `yaml:"test"`
	Build   Stage `yaml:"build"`
	Publish Stage `yaml:"publish"`
}

type Stage struct {
	Enabled      *bool    `yaml:"enabled"`
	PreCommands  []string `yaml:"preCommands"`
	Commands     []string `yaml:"commands"`
	PostCommands []string `yaml:"postCommands"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", path, err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Version != 1 {
		return fmt.Errorf("unsupported config version %d (expected 1)", cfg.Version)
	}
	if cfg.Project.Name == "" {
		return fmt.Errorf("project.name is required")
	}
	if cfg.Coin.Template == "" {
		return fmt.Errorf("coin.template is required")
	}
	if cfg.Jenkins.Credentials.Docker == "" {
		return fmt.Errorf("jenkins.credentials.docker is required")
	}
	return nil
}

// DockerCredentialID — Jenkins Credential ID для registry publish.
func (cfg *Config) DockerCredentialID() string {
	return cfg.Jenkins.Credentials.Docker
}

func (s *Stage) IsEnabled() bool {
	return s.Enabled == nil || *s.Enabled
}

// RuntimeVersion — override из jenkins.runtime, иначе fallback.
func (cfg *Config) RuntimeVersion(key, defaultVersion string) string {
	if v := cfg.Jenkins.Runtime[key]; v != "" {
		return v
	}
	return defaultVersion
}
