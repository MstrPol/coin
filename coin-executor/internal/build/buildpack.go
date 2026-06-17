package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const defaultBuildpackTestImage = "localhost/coin-buildpack-test:discard"

type BuildpackOptions struct {
	Workspace string
	Builder   string
	RunImage  string
	CacheRef  string
	Output    string
	Env       []string
}

func RunBuildpack(opts BuildpackOptions) error {
	if err := ensurePodman(); err != nil {
		return err
	}
	builder := strings.TrimSpace(opts.Builder)
	if builder == "" {
		return fmt.Errorf("buildpack builder is required")
	}

	args, err := packArgs(opts)
	if err != nil {
		return err
	}

	cmd := exec.Command("pack", args...)
	cmd.Dir = opts.Workspace
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PACK_VOLUME_KEY=coin-buildpack")
	fmt.Printf("==> pack %s\n", strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buildpack build: %w", err)
	}
	return nil
}

func packArgs(opts BuildpackOptions) ([]string, error) {
	imageRef, publish, err := parsePackImageRef(opts.Output)
	if err != nil {
		return nil, err
	}
	builder := strings.TrimSpace(opts.Builder)
	args := []string{
		"build", imageRef,
		"--builder", builder,
		"--path", opts.Workspace,
		"--pull-policy", "if-not-present",
		"--trust-builder",
		"--docker-host", "inherit",
		"--network", "host",
	}
	if publish {
		args = append(args, "--publish")
		if opts.CacheRef != "" {
			args = append(args, "--cache-image", opts.CacheRef)
		}
	}
	if runImage := strings.TrimSpace(opts.RunImage); runImage != "" {
		args = append(args, "--run-image", runImage)
	}
	for _, env := range opts.Env {
		env = strings.TrimSpace(env)
		if env == "" {
			continue
		}
		args = append(args, "--env", env)
	}
	return args, nil
}

func parsePackImageRef(output string) (imageRef string, publish bool, err error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return defaultBuildpackTestImage, false, nil
	}
	if !strings.HasPrefix(output, "type=image,") {
		return "", false, fmt.Errorf("unsupported buildpack output %q", output)
	}
	for _, part := range strings.Split(strings.TrimPrefix(output, "type=image,"), ",") {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasPrefix(part, "name="):
			imageRef = strings.TrimPrefix(part, "name=")
		case part == "push=true":
			publish = true
		}
	}
	if imageRef == "" {
		return "", false, fmt.Errorf("image name missing in output %q", output)
	}
	return imageRef, publish, nil
}
