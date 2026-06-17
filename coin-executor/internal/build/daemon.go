package build

import (
	"fmt"
	"os"
	"strings"
)

func ensureBuildkit() error {
	host := strings.TrimSpace(os.Getenv("BUILDKIT_HOST"))
	if host == "" {
		sock := "/tmp/buildkit.sock"
		if _, err := os.Stat(sock); err != nil {
			return fmt.Errorf("BUILDKIT_HOST not set and %s missing; start buildkitd first", sock)
		}
		os.Setenv("BUILDKIT_HOST", "unix://"+sock)
		return nil
	}
	if strings.HasPrefix(host, "unix://") {
		sock := strings.TrimPrefix(host, "unix://")
		if _, err := os.Stat(sock); err != nil {
			return fmt.Errorf("BUILDKIT_HOST=%s but socket missing; start buildkitd first", host)
		}
	}
	return nil
}
