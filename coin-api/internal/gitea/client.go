package gitea

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var ErrTagExists = errors.New("git tag already exists")

type Client struct {
	baseURL      string
	user         string
	password     string
	platformRepo string
	httpClient   *http.Client
}

func NewFromEnv() *Client {
	if envBool("GIT_EXPORT_DISABLED", true) {
		return nil
	}
	base := os.Getenv("GITEA_URL")
	if base == "" {
		base = "http://gitea:3000"
	}
	repo := os.Getenv("GITEA_PLATFORM_REPO")
	if repo == "" {
		repo = "coin/coin-jenkins-agents"
	}
	user := os.Getenv("GITEA_USER")
	if user == "" {
		user = "coin"
	}
	pass := os.Getenv("GITEA_PASSWORD")
	if pass == "" {
		pass = "coin"
	}
	return &Client{
		baseURL:      strings.TrimSuffix(base, "/"),
		user:         user,
		password:     pass,
		platformRepo: repo,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) CreateTag(ctx context.Context, tagName string) error {
	if c == nil {
		return nil
	}
	if tagName == "" {
		return fmt.Errorf("tag name is required")
	}

	owner, repo, err := splitRepo(c.platformRepo)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"tag_name": tagName,
		"target":   "main",
	})
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/tags", c.baseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.user, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusConflict {
		return ErrTagExists
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("gitea create tag %s: %s %s", tagName, resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func splitRepo(full string) (owner, repo string, err error) {
	parts := strings.Split(full, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo %q, want owner/name", full)
	}
	return parts[0], parts[1], nil
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v == "1" || strings.EqualFold(v, "true")
}
