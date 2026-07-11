package build

import (
	"fmt"
	"os"
)

const podmanSocketPath = "/var/run/docker.sock"

func ensurePodman() error {
	if _, err := os.Stat(podmanSocketPath); err != nil {
		return fmt.Errorf("podman socket %s missing; start podman system service in bootstrap", podmanSocketPath)
	}
	if err := os.Setenv("DOCKER_HOST", "unix://"+podmanSocketPath); err != nil {
		return fmt.Errorf("set DOCKER_HOST: %w", err)
	}
	return nil
}
