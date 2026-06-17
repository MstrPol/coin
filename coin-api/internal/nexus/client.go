package nexus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	repository string
	httpClient *http.Client
}

func NewFromEnv() *Client {
	base := os.Getenv("NEXUS_URL")
	if base == "" {
		base = "http://nexus:8081"
	}
	repo := os.Getenv("NEXUS_MAVEN_RELEASES")
	if repo == "" {
		repo = MavenReleases
	}
	return &Client{
		baseURL:    strings.TrimRight(base, "/"),
		repository: repo,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// PointerDoc is the mutable Nexus pointer file (pin → blob).
type PointerDoc struct {
	ResolvedVersion string `json:"resolvedVersion"`
	ManifestHash    string `json:"manifestHash"`
	BlobURL         string `json:"blobUrl"`
}

// UploadManifestBlob stores an immutable manifest snapshot at coin/manifest/{gp}/{version}/{gp}-{version}.json.
func (c *Client) UploadManifestBlob(ctx context.Context, gpName, version string, doc map[string]any) (string, error) {
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	repo := MavenRepoForVersion(version)
	path := MavenRepoPath("coin.manifest", gpName, version, "", "json")
	return c.putRepo(ctx, repo, path, raw, "application/json")
}

// UploadContentArtifact stores GP stage/schema/dockerfile bytes in Nexus.
func (c *Client) UploadContentArtifact(ctx context.Context, gpName, version, artifactKey string, body []byte) (string, error) {
	classifier, ext := ArtifactMavenCoords(artifactKey)
	repo := MavenRepoForVersion(version)
	path := MavenRepoPath("coin.gp.content", gpName, version, classifier, ext)
	contentType := "application/octet-stream"
	if ext == "json" {
		contentType = "application/json"
	} else if ext == "sh" {
		contentType = "text/x-shellscript"
	}
	return c.putRepo(ctx, repo, path, body, contentType)
}

// PutPointer updates the mutable pointer for a GP pin (e.g. "=1.0.0").
func (c *Client) PutPointer(ctx context.Context, gpName, pin string, pointer PointerDoc) (string, error) {
	raw, err := json.MarshalIndent(pointer, "", "  ")
	if err != nil {
		return "", err
	}
	classifier := "pin-" + encodePinForPath(pin)
	path := MavenRepoPath("coin.manifest", gpName, "metadata", classifier, "json")
	return c.putRepo(ctx, MavenSnapshots, path, raw, "application/json")
}

// PublishManifest uploads the immutable blob and updates the exact-pin pointer.
func (c *Client) PublishManifest(ctx context.Context, gpName, version, pin string, doc map[string]any, manifestHash string) (string, error) {
	blobURL, err := c.UploadManifestBlob(ctx, gpName, version, doc)
	if err != nil {
		return "", fmt.Errorf("upload blob: %w", err)
	}
	ptrURL, err := c.PutPointer(ctx, gpName, pin, PointerDoc{
		ResolvedVersion: version,
		ManifestHash:    manifestHash,
		BlobURL:         blobURL,
	})
	if err != nil {
		return "", fmt.Errorf("update pointer: %w", err)
	}
	return ptrURL, nil
}

// UploadManifest is deprecated: use PublishManifest (blob + pointer).
func (c *Client) UploadManifest(ctx context.Context, gpName, version string, doc map[string]any) (string, error) {
	hash, _ := doc["manifestHash"].(string)
	if hash == "" {
		return "", fmt.Errorf("manifestHash missing")
	}
	return c.PublishManifest(ctx, gpName, version, "="+version, doc, hash)
}

func (c *Client) put(ctx context.Context, repoPath string, body []byte, contentType string) (string, error) {
	return c.putRepo(ctx, c.repository, repoPath, body, contentType)
}

func (c *Client) putRepo(ctx context.Context, repo, repoPath string, body []byte, contentType string) (string, error) {
	target := fmt.Sprintf("%s/repository/%s/%s", c.baseURL, repo, repoPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(os.Getenv("NEXUS_ADMIN_USER"), os.Getenv("NEXUS_ADMIN_PASSWORD"))
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		msg := string(respBody)
		// Nexus Docker connector may put the reason in the status line, not the body.
		if ImmutableConflict(resp.StatusCode, msg) || ImmutableConflict(resp.StatusCode, resp.Status) {
			return target, nil
		}
		return "", fmt.Errorf("nexus upload %s: %s", resp.Status, msg)
	}
	return target, nil
}

func encodePinForPath(pin string) string {
	r := strings.NewReplacer("=", "%3D", "~", "%7E", "^", "%5E")
	return r.Replace(pin)
}
