package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultPath = ".coin/config.yaml"

// Config — полная схема .coin/config.yaml.
//
// Секция agent: читается Jenkins (coin-lib) для подготовки динамического
// агента и credentials. Всё остальное читает только coin CLI.
type Config struct {
	Version   int       `yaml:"version"`
	Coin      CoinMeta  `yaml:"coin"`
	Agent     Agent     `yaml:"agent"`     // → Jenkins (coin-lib)
	Project   Project   `yaml:"project"`   // → coin CLI
	Container Container `yaml:"container"` // → coin CLI
	Pipeline  Pipeline  `yaml:"pipeline"`  // → coin CLI
}

// Agent — секция для Jenkins.
// coin-lib читает только эти поля; разработчик видит явную границу в файле.
type Agent struct {
	Stack           string            `yaml:"stack"`           // стек (python-uv, go, java-maven, …)
	Runtime         map[string]string `yaml:"runtime"`         // версии toolchain (python: "3.13")
	PublishRegistry string            `yaml:"publishRegistry"` // Jenkins Credential ID для публикации
}

type CoinMeta struct {
	Template        string     `yaml:"template"`
	TemplateVersion string `yaml:"templateVersion"`
}

type Project struct {
	Name string `yaml:"name"`
}

type Container struct {
	Port    int      `yaml:"port"`
	Command []string `yaml:"command"`
}

type Pipeline struct {
	Test    Stage        `yaml:"test"`
	Build   BuildStage   `yaml:"build"`
	Publish PublishStage `yaml:"publish"`
}

type Stage struct {
	Enabled      *bool    `yaml:"enabled"`
	PreCommands  []string `yaml:"preCommands"`
	Commands     []string `yaml:"commands"`
	PostCommands []string `yaml:"postCommands"`
}

type BuildStage struct {
	Stage              `yaml:",inline"`
	Target             string `yaml:"target"`
	DockerfileTemplate string `yaml:"dockerfileTemplate"`
}

type PublishStage struct {
	Stage `yaml:",inline"`
	When  string `yaml:"when"`
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
	if cfg.Agent.Stack == "" {
		return fmt.Errorf("agent.stack is required")
	}

	allowed := map[string]bool{
		"python-uv": true, "python-pip": true,
		"java-maven": true, "java-gradle": true,
		"go": true, "node": true,
	}
	if !allowed[cfg.Agent.Stack] {
		return fmt.Errorf("unknown agent.stack %q", cfg.Agent.Stack)
	}

	return nil
}

// Stack — стек из agent-секции; используется как coin CLI, так и Jenkins.
func (cfg *Config) Stack() string {
	return cfg.Agent.Stack
}

// RuntimeVersion — версия toolchain для указанного ключа.
func (cfg *Config) RuntimeVersion(key, defaultVersion string) string {
	if v := cfg.Agent.Runtime[key]; v != "" {
		return v
	}
	return defaultVersion
}

func (s *Stage) IsEnabled() bool {
	return s.Enabled == nil || *s.Enabled
}

func (cfg *Config) BuildTarget() string {
	if cfg.Pipeline.Build.Target != "" {
		return cfg.Pipeline.Build.Target
	}
	return "package"
}
