package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Options struct {
	Workspace  string
	Dockerfile string
	Target     string
	CacheRef   string
	Output     string
}

func RunTarget(opts Options) error {
	if err := ensurePodman(); err == nil {
		return RunPodmanTarget(opts)
	}
	if err := ensureBuildkit(); err != nil {
		return err
	}
	dockerfilePath := opts.Dockerfile
	if !filepath.IsAbs(dockerfilePath) {
		dockerfilePath = filepath.Join(opts.Workspace, dockerfilePath)
	}
	dockerfileDir := filepath.Dir(dockerfilePath)
	filename := filepath.Base(dockerfilePath)

	args := buildkitArgs(opts.Workspace, dockerfileDir, filename, opts)

	cmd := exec.Command("buildctl", args...)
	cmd.Dir = opts.Workspace
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	fmt.Printf("==> buildctl %s\n", strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buildctl target %s: %w", opts.Target, err)
	}
	return nil
}

func buildkitArgs(workspace, dockerfileDir, filename string, opts Options) []string {
	args := []string{
		"build",
		"--frontend", "dockerfile.v0",
		"--local", "context=" + workspace,
		"--local", "dockerfile=" + dockerfileDir,
		"--opt", "filename=" + filename,
	}
	if opts.Target != "" {
		args = append(args, "--opt", "target="+opts.Target)
	}
	if platform := nativePlatform(); platform != "" {
		args = append(args, "--opt", "platform="+platform)
	}
	if opts.CacheRef != "" {
		args = append(args, "--import-cache", "type=registry,ref="+opts.CacheRef)
		if opts.Output != "" {
			args = append(args, "--export-cache", "type=registry,ref="+opts.CacheRef, "mode=max")
		}
	}
	if opts.Output != "" {
		args = append(args, "--output", opts.Output)
	}
	return args
}

func nativePlatform() string {
	switch runtime.GOARCH {
	case "arm64":
		return "linux/arm64"
	case "amd64":
		return "linux/amd64"
	default:
		return ""
	}
}
