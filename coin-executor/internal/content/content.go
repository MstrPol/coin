package content

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"coin.local/coin-executor/internal/manifest"
)

const seedLocalRef = "seed-local"

// EnsureRoot returns a local content tree root when required by legacy gitRef manifests.
func EnsureRoot(m *manifest.Manifest) (string, error) {
	if dir := strings.TrimSpace(os.Getenv("COIN_CONTENT_DIR")); dir != "" {
		return dir, nil
	}
	if dir := strings.TrimSpace(os.Getenv("COIN_PLATFORM_DIR")); dir != "" {
		return dir, nil
	}
	if m.URLRefsOnly() {
		return "", nil
	}

	refs := m.ContentGitRefs()
	if len(refs) == 0 {
		return "", fmt.Errorf("manifest has no content url or gitRef")
	}
	if len(refs) > 1 {
		return "", fmt.Errorf("multiple gitRef in manifest (%v); set COIN_CONTENT_DIR", refs)
	}
	gitRef := refs[0]
	if gitRef == seedLocalRef {
		return "", fmt.Errorf("gitRef %q requires COIN_CONTENT_DIR or COIN_PLATFORM_DIR", gitRef)
	}

	dest := strings.TrimSpace(os.Getenv("COIN_CONTENT_CHECKOUT_DIR"))
	if dest == "" {
		dest = filepath.Join(".coin", "content")
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}

	repoURL := strings.TrimSpace(os.Getenv("COIN_PLATFORM_GIT_URL"))
	if repoURL == "" {
		repoURL = "http://gitea:3000/coin/coin-platform.git"
	}

	if err := os.RemoveAll(dest); err != nil {
		return "", err
	}
	if err := shallowCheckout(repoURL, gitRef, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func Materialize(dest string, contentRoot string, ref manifest.ContentRef) error {
	data, err := readContent(contentRoot, ref)
	if err != nil {
		return err
	}
	if want := strings.TrimSpace(ref.SHA256); want != "" {
		sum := sha256.Sum256(data)
		got := "sha256:" + hex.EncodeToString(sum[:])
		if got != want {
			label := ref.URL
			if label == "" {
				label = ref.Path
			}
			return fmt.Errorf("content %s sha256 mismatch: got %s want %s", label, got, want)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if strings.HasSuffix(dest, ".sh") {
		mode = 0o755
	}
	return os.WriteFile(dest, data, mode)
}

func readContent(contentRoot string, ref manifest.ContentRef) ([]byte, error) {
	if url := strings.TrimSpace(ref.URL); url != "" {
		return fetchURL(url)
	}
	if path := strings.TrimSpace(ref.Path); path != "" {
		return os.ReadFile(filepath.Join(contentRoot, path))
	}
	return nil, fmt.Errorf("content ref missing url and path")
}

func fetchURL(rawURL string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch %s: %s", rawURL, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func shallowCheckout(repoURL, gitRef, dest string) error {
	if err := runGit("", "clone", "--depth", "1", repoURL, dest); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	if gitRef == "main" || gitRef == "master" {
		return nil
	}
	if err := runGit(dest, "fetch", "--depth", "1", "origin", gitRef); err != nil {
		return fmt.Errorf("git fetch %s: %w", gitRef, err)
	}
	if err := runGit(dest, "checkout", "FETCH_HEAD"); err != nil {
		return fmt.Errorf("git checkout %s: %w", gitRef, err)
	}
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GoldenPathsDir(contentRoot string) string {
	if contentRoot == "" {
		return ""
	}
	return filepath.Join(contentRoot, "content", "golden-paths")
}
