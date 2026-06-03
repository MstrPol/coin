package goldenpaths

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const cacheDir = ".coin/cache/golden-paths"

// FetchFromURL скачивает catalog (tar.gz или распакованная директория по file://).
func FetchFromURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("COIN_GOLDEN_PATHS_URL is required for nexus source")
	}
	if strings.HasPrefix(rawURL, "file://") {
		return strings.TrimPrefix(rawURL, "file://"), nil
	}

	sum := sha256.Sum256([]byte(rawURL))
	dest := filepath.Join(cacheDir, hex.EncodeToString(sum[:8]))
	if _, err := os.Stat(filepath.Join(dest, catalogFile)); err == nil {
		return dest, nil
	}

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", err
	}

	resp, err := http.Get(rawURL) //nolint:gosec // URL from platform env
	if err != nil {
		return "", fmt.Errorf("fetch templates: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch templates: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "coin-golden-paths-*.tar.gz")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return "", err
	}
	tmp.Close()

	if err := extractTarGz(tmpPath, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func extractTarGz(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, hdr.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid tar path: %s", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
}
