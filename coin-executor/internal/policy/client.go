package policy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Result struct {
	Warning string
}

// CheckResolvedVersion calls coin-api policy-check for manifest.goldenPath.version.
func CheckResolvedVersion(gpName, resolvedVersion string) (Result, error) {
	base := strings.TrimRight(os.Getenv("COIN_API_URL"), "/")
	if base == "" {
		base = "http://coin-api:8090"
	}
	token := os.Getenv("COIN_API_TOKEN")
	if token == "" {
		return Result{}, fmt.Errorf("COIN_API_TOKEN is required for GP policy check")
	}

	u, err := url.Parse(fmt.Sprintf("%s/v1/golden-paths/%s/policy-check", base, gpName))
	if err != nil {
		return Result{}, err
	}
	q := u.Query()
	q.Set("version", resolvedVersion)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusForbidden {
		var errBody struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(body, &errBody)
		if errBody.Error != "" {
			return Result{}, fmt.Errorf("%s", errBody.Error)
		}
		return Result{}, fmt.Errorf("GP version %s below catalog minimum", resolvedVersion)
	}
	if resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("policy-check %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var ok struct {
		Warning string `json:"warning"`
	}
	if err := json.Unmarshal(body, &ok); err != nil {
		return Result{}, err
	}
	return Result{Warning: ok.Warning}, nil
}
