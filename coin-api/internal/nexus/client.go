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
	repo := os.Getenv("NEXUS_MANIFEST_REPO")
	if repo == "" {
		repo = "coin-manifests"
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

// UploadManifestBlob stores an immutable manifest snapshot keyed by manifestHash.
func (c *Client) UploadManifestBlob(ctx context.Context, manifestHash string, doc map[string]any) (string, error) {
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("manifest/blobs/%s.json", blobFileName(manifestHash))
	return c.put(ctx, path, raw, "application/json")
}

// UploadContentArtifact stores GP stage/schema/dockerfile bytes in Nexus.
func (c *Client) UploadContentArtifact(ctx context.Context, gpName, version, artifactKey string, body []byte) (string, error) {
	path := fmt.Sprintf("content/%s/%s/%s", gpName, version, artifactKey)
	contentType := "application/octet-stream"
	if strings.HasSuffix(artifactKey, ".json") {
		contentType = "application/json"
	} else if strings.HasSuffix(artifactKey, ".sh") {
		contentType = "text/x-shellscript"
	} else if artifactKey == "Dockerfile" {
		contentType = "text/plain"
	}
	return c.put(ctx, path, body, contentType)
}

// PutPointer updates the mutable pointer for a GP pin (e.g. "=1.0.0").
func (c *Client) PutPointer(ctx context.Context, gpName, pin string, pointer PointerDoc) (string, error) {
	raw, err := json.MarshalIndent(pointer, "", "  ")
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("pointers/%s/%s.json", gpName, encodePinForPath(pin))
	return c.put(ctx, path, raw, "application/json")
}

// PublishManifest uploads the immutable blob and updates the exact-pin pointer.
func (c *Client) PublishManifest(ctx context.Context, gpName, version, pin string, doc map[string]any, manifestHash string) (string, error) {
	blobURL, err := c.UploadManifestBlob(ctx, manifestHash, doc)
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
	target := fmt.Sprintf("%s/repository/%s/%s", c.baseURL, c.repository, repoPath)
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
		return "", fmt.Errorf("nexus upload %s: %s", resp.Status, string(respBody))
	}
	return target, nil
}

func blobFileName(manifestHash string) string {
	return strings.TrimPrefix(manifestHash, "sha256:")
}

func encodePinForPath(pin string) string {
	r := strings.NewReplacer("=", "%3D", "~", "%7E", "^", "%5E")
	return r.Replace(pin)
}
