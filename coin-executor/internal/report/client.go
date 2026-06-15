package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/manifest"
	"coin.local/coin-executor/internal/outputs"
)

type Payload struct {
	Project         string `json:"project"`
	GroupID         string `json:"groupId,omitempty"`
	ArtifactID      string `json:"artifactId,omitempty"`
	GoldenPath      string `json:"goldenPath"`
	Version         string `json:"version"`
	ConfigVersion   string `json:"configVersion,omitempty"`
	Branch          string `json:"branch,omitempty"`
	BuildURL        string `json:"buildUrl,omitempty"`
	Result          string `json:"result"`
	ManifestHash    string `json:"manifestHash,omitempty"`
	GitURL          string `json:"gitUrl,omitempty"`
	Channel         string `json:"channel,omitempty"`
	RequestedPin    string `json:"requestedPin,omitempty"`
	FailedStage     string `json:"failedStage,omitempty"`
	ResolvedVersion string          `json:"resolvedVersion,omitempty"`
	Outputs         []outputs.Entry `json:"outputs,omitempty"`
}

func Submit(projectPath, manifestPath, buildURL, result string) error {
	cfg, err := config.Load(projectPath)
	if err != nil {
		return err
	}
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return err
	}

	wd, _ := os.Getwd()
	reportOutputs, _ := outputs.Load(wd)

	payload := Payload{
		Project:         cfg.Project.Name,
		GroupID:         cfg.Project.GroupID,
		ArtifactID:      cfg.Project.ArtifactID,
		GoldenPath:      m.GoldenPath.Name,
		Version:         m.GoldenPath.Version,
		ConfigVersion:   cfg.Coin.Version,
		Branch:          strings.TrimSpace(os.Getenv("GIT_BRANCH")),
		BuildURL:        buildURL,
		Result:          result,
		ManifestHash:    m.ManifestHash,
		GitURL:          strings.TrimSpace(os.Getenv("GIT_URL")),
		Channel:         strings.TrimSpace(os.Getenv("COIN_CHANNEL")),
		RequestedPin:    strings.TrimSpace(os.Getenv("COIN_REQUESTED_PIN")),
		FailedStage:     strings.TrimSpace(os.Getenv("COIN_FAILED_STAGE")),
		ResolvedVersion: m.GoldenPath.Version,
		Outputs:         reportOutputs,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	base := strings.TrimSpace(os.Getenv("COIN_API_URL"))
	if base == "" {
		base = "http://coin-api:8090"
	}
	url := strings.TrimRight(base, "/") + "/v1/builds/report"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := strings.TrimSpace(os.Getenv("COIN_API_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post report: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("post report: HTTP %s", resp.Status)
	}
	return nil
}
