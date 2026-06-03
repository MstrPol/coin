package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"coin.local/coin-cli/internal/config"
	"coin.local/coin-cli/internal/dockerfile"
	"coin.local/coin-cli/internal/goldenpaths"
	"coin.local/coin-cli/internal/versioning"
)

type Executor struct {
	cfg    *config.Config
	bundle *goldenpaths.Bundle
}

func New(cfg *config.Config, bundle *goldenpaths.Bundle) *Executor {
	return &Executor{cfg: cfg, bundle: bundle}
}

func (e *Executor) Run(stage string) error {
	ver, err := versioning.Compute("")
	if err != nil {
		return err
	}

	env := e.buildEnv(ver)

	if !e.stageEnabled(stage) {
		fmt.Printf("==> stage %s disabled, skipping\n", stage)
		return nil
	}

	var stg config.Stage
	switch stage {
	case "test":
		stg = e.cfg.Pipeline.Test
	case "build":
		stg = e.cfg.Pipeline.Build
	case "publish":
		stg = e.cfg.Pipeline.Publish
	default:
		return fmt.Errorf("unknown stage: %s", stage)
	}

	if err := runCommands(stg.PreCommands, env); err != nil {
		return fmt.Errorf("preCommands failed: %w", err)
	}

	if len(stg.Commands) > 0 {
		if err := runCommands(stg.Commands, env); err != nil {
			return fmt.Errorf("commands failed: %w", err)
		}
	} else {
		if err := e.runStandardScript(stage, env); err != nil {
			return err
		}
	}

	if err := runCommands(stg.PostCommands, env); err != nil {
		return fmt.Errorf("postCommands failed: %w", err)
	}

	return nil
}

func (e *Executor) stageEnabled(stage string) bool {
	switch stage {
	case "test":
		return e.bundle.StageEnabled("test", e.cfg.Pipeline.Test.Enabled)
	case "build":
		return e.bundle.StageEnabled("build", e.cfg.Pipeline.Build.Enabled)
	case "publish":
		return e.bundle.StageEnabled("publish", e.cfg.Pipeline.Publish.Enabled)
	default:
		return true
	}
}

func (e *Executor) runStandardScript(stage string, env []string) error {
	if stage == "build" && e.bundle.BuildType() == "container" {
		df, err := dockerfile.Render(e.cfg, e.bundle)
		if err != nil {
			return err
		}
		env = append(env, "COIN_DOCKERFILE="+df)
	}

	scriptContent, err := e.bundle.Script(stage)
	if err != nil {
		return fmt.Errorf("no standard script for %s/%s stage %s: %w",
			e.bundle.Name, e.bundle.Version, stage, err)
	}

	tmp, err := os.CreateTemp("", "coin-*.sh")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(scriptContent); err != nil {
		return err
	}
	tmp.Close()

	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return err
	}

	return runShell(tmp.Name(), env)
}

func (e *Executor) buildEnv(ver *versioning.Result) []string {
	target := e.bundle.BuildType()
	name := e.cfg.Project.Name
	registry := os.Getenv("COIN_REGISTRY_PREFIX")

	ref := name + ":" + ver.ImageTag
	if registry != "" {
		ref = strings.TrimRight(registry, "/") + "/" + ref
	}

	base := os.Environ()
	extra := []string{
		"COIN_VERSION=" + ver.Version,
		"COIN_VERSION_SOURCE=" + ver.Source,
		"COIN_IMAGE_TAG=" + ver.ImageTag,
		"COIN_IMAGE_NAME=" + name,
		"COIN_IMAGE_REF=" + ref,
		"COIN_BUILD_TARGET=" + target,
		"COIN_TEMPLATE=" + e.bundle.Name,
		"COIN_TEMPLATE_VERSION=" + e.bundle.Version,
	}
	return append(base, extra...)
}

func runCommands(cmds []string, env []string) error {
	for _, c := range cmds {
		if err := runShell(c, env); err != nil {
			return err
		}
	}
	return nil
}

func runShell(command string, env []string) error {
	cmd := exec.Command("bash", "-euo", "pipefail", "-c", command) //nolint:gosec
	if strings.HasPrefix(command, "/") || strings.HasSuffix(command, ".sh") {
		cmd = exec.Command(command) //nolint:gosec
	}
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
