package scanner

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type GiteaClient struct {
	baseURL    string
	user       string
	password   string
	org        string
	httpClient *http.Client
}

func NewGiteaFromEnv() *GiteaClient {
	base := os.Getenv("GITEA_URL")
	if base == "" {
		base = "http://gitea:3000"
	}
	org := os.Getenv("GITEA_ORG")
	if org == "" {
		org = "coin"
	}
	user := os.Getenv("GITEA_USER")
	if user == "" {
		user = "coin"
	}
	pass := os.Getenv("GITEA_PASSWORD")
	if pass == "" {
		pass = "coin"
	}
	return &GiteaClient{
		baseURL:    strings.TrimSuffix(base, "/"),
		user:       user,
		password:   pass,
		org:        org,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type Repo struct {
	FullName      string
	Name          string
	DefaultBranch string
	HTMLURL       string
}

type giteaRepo struct {
	FullName      string `json:"full_name"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	HTMLURL       string `json:"html_url"`
}

func (c *GiteaClient) ListRepos(ctx context.Context) ([]Repo, error) {
	url := fmt.Sprintf("%s/api/v1/users/%s/repos?limit=100", c.baseURL, c.org)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list repos: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var raw []giteaRepo
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]Repo, 0, len(raw))
	for _, r := range raw {
		branch := r.DefaultBranch
		if branch == "" {
			branch = "main"
		}
		out = append(out, Repo{
			FullName:      r.FullName,
			Name:          r.Name,
			DefaultBranch: branch,
			HTMLURL:       r.HTMLURL,
		})
	}
	return out, nil
}

func (c *GiteaClient) BranchSHA(ctx context.Context, owner, repo, branch string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/branches/%s", c.baseURL, owner, repo, branch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("branch sha: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Commit struct {
			ID string `json:"id"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.Commit.ID == "" {
		return "", fmt.Errorf("empty sha for %s/%s@%s", owner, repo, branch)
	}
	return payload.Commit.ID, nil
}

func (c *GiteaClient) RawFile(ctx context.Context, owner, repo, branch, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/raw/%s/%s", c.baseURL, owner, repo, branch, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNoConfig
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("raw file: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return io.ReadAll(resp.Body)
}

// ContentsFile fallback when raw endpoint unavailable.
func (c *GiteaClient) ContentsFile(ctx context.Context, owner, repo, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/contents/%s", c.baseURL, owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNoConfig
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("contents: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Encoding != "base64" {
		return nil, fmt.Errorf("unsupported encoding %q", payload.Encoding)
	}
	return base64.StdEncoding.DecodeString(strings.TrimSpace(payload.Content))
}

func (c *GiteaClient) CloneURL(fullName string) string {
	return fmt.Sprintf("%s/%s.git", c.baseURL, fullName)
}

func SplitOwnerRepo(full string) (owner, name string, err error) {
	parts := strings.Split(full, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo %q", full)
	}
	return parts[0], parts[1], nil
}
