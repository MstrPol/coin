package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunPodmanTarget(opts Options) error {
	if err := ensurePodman(); err != nil {
		return err
	}
	dockerfilePath := opts.Dockerfile
	if !filepath.IsAbs(dockerfilePath) {
		dockerfilePath = filepath.Join(opts.Workspace, dockerfilePath)
	}

	args := []string{"build", "-f", dockerfilePath}
	if target := strings.TrimSpace(opts.Target); target != "" {
		args = append(args, "--target", target)
	}

	imageRef, push, err := parseImageOutput(opts.Output)
	if err != nil {
		return err
	}
	if imageRef != "" {
		args = append(args, "-t", imageRef)
	}
	args = append(args, opts.Workspace)

	cmd := exec.Command("podman", args...)
	cmd.Dir = opts.Workspace
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	fmt.Printf("==> podman %s\n", strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman build target %s: %w", opts.Target, err)
	}

	if push && imageRef != "" {
		pushCmd := exec.Command("podman", "push", imageRef)
		pushCmd.Dir = opts.Workspace
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		pushCmd.Env = os.Environ()
		fmt.Printf("==> podman push %s\n", imageRef)
		if err := pushCmd.Run(); err != nil {
			return fmt.Errorf("podman push %s: %w", imageRef, err)
		}
	}
	return nil
}

func parseImageOutput(output string) (imageRef string, push bool, err error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return "", false, nil
	}
	if !strings.HasPrefix(output, "type=image,") {
		return "", false, fmt.Errorf("unsupported podman output %q", output)
	}
	for _, part := range strings.Split(strings.TrimPrefix(output, "type=image,"), ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			return "", false, fmt.Errorf("invalid podman output segment %q", part)
		}
		switch key {
		case "name":
			imageRef = val
		case "push":
			push = val == "true"
		}
	}
	if imageRef == "" {
		return "", false, fmt.Errorf("podman output missing image name: %q", output)
	}
	return imageRef, push, nil
}
