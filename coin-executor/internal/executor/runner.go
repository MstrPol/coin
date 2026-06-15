package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/content"
	"coin.local/coin-executor/internal/deliverables"
	"coin.local/coin-executor/internal/manifest"
)

type RunOptions struct {
	Stage string
}

type Runner struct {
	Workspace string
}

func (r Runner) Run(cfg *config.Config, m *manifest.Manifest, opts RunOptions) error {
	items := cfg.NormalizedDeliverables()
	if err := deliverables.Validate(items, m.AllowedDeliverableTypes()); err != nil {
		return err
	}
	if err := deliverables.WriteState(r.Workspace, items); err != nil {
		return err
	}

	contentRoot, err := content.EnsureRoot(m)
	if err != nil {
		return err
	}

	goldenPaths := content.GoldenPathsDir(contentRoot)
	dockerfile := filepath.Join(r.Workspace, ".coin", "Dockerfile")
	if err := content.Materialize(dockerfile, contentRoot, m.DockerfileTemplate); err != nil {
		return fmt.Errorf("dockerfile: %w", err)
	}
	manifestPath := filepath.Join(r.Workspace, ".coin", "manifest.json")

	for _, stage := range m.Pipeline.Stages {
		if opts.Stage != "" && stage.Name != opts.Stage {
			continue
		}
		if !shouldRunStage(stage) {
			fmt.Printf("==> skip stage %s (when=%s)\n", stage.Name, stage.When)
			continue
		}
		scriptPath := filepath.Join(r.Workspace, ".coin", "stage-"+stage.Name+".sh")
		if err := content.Materialize(scriptPath, contentRoot, stage.Script); err != nil {
			return err
		}

		cmd := exec.Command("/bin/bash", scriptPath)
		cmd.Dir = r.Workspace
		env := []string{
			"COIN_DOCKERFILE=" + dockerfile,
			"COIN_MANIFEST_PATH=" + manifestPath,
			"COIN_GP=" + cfg.Coin.GoldenPath,
			"COIN_GP_VERSION=" + cfg.Coin.Version,
		}
		if contentRoot != "" {
			env = append(env,
				"COIN_CONTENT_DIR="+contentRoot,
				"COIN_PLATFORM_DIR="+contentRoot,
				"COIN_GOLDEN_PATHS_DIR="+goldenPaths,
			)
		}
		cmd.Env = append(os.Environ(), env...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("==> stage %s\n", stage.Name)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("stage %s: %w", stage.Name, err)
		}
	}
	return nil
}

func shouldRunStage(stage manifest.Stage) bool {
	switch stage.When {
	case "", "always":
		return true
	case "tag":
		if tag := strings.TrimSpace(os.Getenv("TAG_NAME")); tag != "" {
			return true
		}
		return strings.TrimSpace(os.Getenv("GIT_TAG_NAME")) != ""
	default:
		return true
	}
}
