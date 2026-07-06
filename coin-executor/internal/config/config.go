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
	ArtifactID string `yaml:"artifactId"`
	GroupID    string `yaml:"groupId"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := rejectUnsupportedFields(data); err != nil {
		return nil, err
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
	if cfg.Project.ArtifactID == "" {
		return fmt.Errorf("project.artifactId is required (config v2)")
	}
	if cfg.Project.GroupID == "" {
		return fmt.Errorf("project.groupId is required (config v2)")
	}
	if cfg.Jenkins.Credentials.Docker == "" {
		return fmt.Errorf("jenkins.credentials.docker is required")
	}
	return nil
}

func rejectUnsupportedFields(data []byte) error {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	if len(doc.Content) == 0 {
		return nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("config root must be a mapping")
	}

	allowedTop := map[string]struct{}{
		"coin":    {},
		"jenkins": {},
		"project": {},
	}
	for i := 0; i < len(root.Content); i += 2 {
		key := root.Content[i].Value
		if _, ok := allowedTop[key]; !ok {
			return fmt.Errorf("%s is not supported in product config v2", key)
		}
		if key == "coin" {
			if err := rejectUnsupportedMappingFields(root.Content[i+1], "coin", map[string]struct{}{
				"goldenPath": {},
				"version":    {},
			}); err != nil {
				return err
			}
		}
		if key == "project" {
			if err := rejectUnsupportedMappingFields(root.Content[i+1], "project", map[string]struct{}{
				"name":       {},
				"groupId":    {},
				"artifactId": {},
			}); err != nil {
				return err
			}
		}
		if key == "jenkins" {
			if err := rejectUnsupportedMappingFields(root.Content[i+1], "jenkins", map[string]struct{}{
				"credentials": {},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func rejectUnsupportedMappingFields(node *yaml.Node, prefix string, allowed map[string]struct{}) error {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("%s.%s is not supported in product config v2", prefix, key)
		}
	}
	return nil
}
