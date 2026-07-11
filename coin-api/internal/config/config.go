package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr            string
	DatabaseURL     string
	AuthDisabled    bool
	APIToken        string
	AdminAPIKey     string
	PublisherAPIKey string
	ReaderAPIKey    string
	OIDCEnabled     bool
	OIDCIssuerURL   string
	OIDCAudience    string
	OIDCRolesClaim  string
	LogLevel        string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:            envOr("COIN_API_ADDR", ":8090"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		AuthDisabled:    envBool("AUTH_DISABLED", false),
		APIToken:        os.Getenv("COIN_API_TOKEN"),
		AdminAPIKey:     os.Getenv("COIN_ADMIN_API_KEY"),
		PublisherAPIKey: os.Getenv("COIN_PUBLISHER_API_KEY"),
		ReaderAPIKey:    os.Getenv("COIN_READER_API_KEY"),
		OIDCEnabled:     envBool("OIDC_ENABLED", false),
		OIDCIssuerURL:   os.Getenv("OIDC_ISSUER_URL"),
		OIDCAudience:    os.Getenv("OIDC_AUDIENCE"),
		OIDCRolesClaim:  envOr("OIDC_ROLES_CLAIM", "roles"),
		LogLevel:        envOr("LOG_LEVEL", "info"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if !cfg.AuthDisabled && cfg.APIToken == "" {
		return Config{}, fmt.Errorf("COIN_API_TOKEN is required when AUTH_DISABLED=false")
	}
	if !cfg.AuthDisabled && !cfg.OIDCEnabled && cfg.AdminAPIKey == "" {
		return Config{}, fmt.Errorf("COIN_ADMIN_API_KEY is required when AUTH_DISABLED=false and OIDC_ENABLED=false")
	}
	if !cfg.AuthDisabled && cfg.OIDCEnabled && cfg.OIDCIssuerURL == "" {
		return Config{}, fmt.Errorf("OIDC_ISSUER_URL is required when OIDC_ENABLED=true")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
