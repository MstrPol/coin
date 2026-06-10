package bootstrap

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"coin.local/coin-executor/internal/manifest"
)

// DownloadExecutor fetches coin-executor binary from manifest URL into dest.
func DownloadExecutor(m *manifest.Manifest, dest string) error {
	if m.Executor.URL == "" {
		return fmt.Errorf("manifest executor.url is empty")
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(m.Executor.URL)
	if err != nil {
		return fmt.Errorf("download executor: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("download executor: HTTP %s", resp.Status)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), "coin-executor-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, hasher), resp.Body); err != nil {
		tmp.Close()
		return fmt.Errorf("write executor: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if want := strings.TrimSpace(m.Executor.SHA256); want != "" {
		got := "sha256:" + hex.EncodeToString(hasher.Sum(nil))
		if got != want {
			return fmt.Errorf("executor sha256 mismatch: got %s want %s", got, want)
		}
	}

	if err := os.Rename(tmpPath, dest); err != nil {
		return err
	}
	return os.Chmod(dest, 0o755)
}
