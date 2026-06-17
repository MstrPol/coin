package executor

import (
	"fmt"
	"os"
	"strings"

	"coin.local/coin-executor/internal/build"
	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/deliverables"
	"coin.local/coin-executor/internal/manifest"
	"coin.local/coin-executor/internal/outputs"
	"coin.local/coin-executor/internal/publish"
	"coin.local/coin-executor/internal/validate"
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

	for _, stage := range m.Pipeline.Stages {
		key := stage.Key()
		if opts.Stage != "" && key != opts.Stage {
			continue
		}
		// Jenkins/coin-executor --stage publish must not be gated by manifest when=tag.
		if opts.Stage == "" && !shouldRunStage(stage) {
			fmt.Printf("==> skip stage %s (when=%s)\n", key, stage.When)
			continue
		}
		fmt.Printf("==> stage %s\n", key)
		if err := r.runStage(cfg, m, key); err != nil {
			return fmt.Errorf("stage %s: %w", key, err)
		}
	}
	return nil
}

func (r Runner) runStage(cfg *config.Config, m *manifest.Manifest, stageID string) error {
	switch stageID {
	case "validate":
		return validate.Project(cfg, m)
	case "test":
		return r.runTest(m)
	case "build":
		return r.runBuild(cfg, m)
	case "publish":
		return r.runPublish(cfg, m)
	default:
		return fmt.Errorf("unknown stage %q", stageID)
	}
}

func (r Runner) runTest(m *manifest.Manifest) error {
	switch m.Build.Engine {
	case "buildkit":
		return r.runBuildkitTarget(m, m.BuildkitTarget("test"))
	case "buildpack":
		return r.runBuildpackTest(m)
	case "dockerfile":
		target := ""
		if m.Build.Dockerfile != nil {
			target = m.Build.Dockerfile.TestTarget
		}
		if target == "" {
			fmt.Println("==> skip test (dockerfile engine: testTarget not set)")
			return nil
		}
		return r.runDockerfileEngineTarget(m, target)
	default:
		return fmt.Errorf("unsupported build engine %q for test stage", m.Build.Engine)
	}
}

func buildpackGoEnv(extra ...string) []string {
	env := []string{
		"BP_GO_VERSION=1.25.8",
		"BP_GO_BUILD_TARGETS=./...",
	}
	return append(env, extra...)
}

func (r Runner) runBuildpackTest(m *manifest.Manifest) error {
	cfg := m.Build.Buildpack
	if cfg == nil {
		return fmt.Errorf("manifest build.buildpack is required")
	}
	return build.RunBuildpack(build.BuildpackOptions{
		Workspace: r.Workspace,
		Builder:   cfg.Builder,
		RunImage:  cfg.RunImage,
		CacheRef:  cfg.CacheRef,
		Env:       buildpackGoEnv("BP_RUN_TESTS=true"),
	})
}

func (r Runner) runBuildkitTarget(m *manifest.Manifest, target string) error {
	if m.Build.Engine != "buildkit" {
		return fmt.Errorf("build engine %q does not support buildkit targets", m.Build.Engine)
	}
	if err := r.materializeContainerfile(m); err != nil {
		return err
	}
	return build.RunTarget(build.Options{
		Workspace:  r.Workspace,
		Dockerfile: m.Build.Buildkit.Dockerfile,
		Target:     target,
		CacheRef:   m.Build.Buildkit.CacheRef,
	})
}

func (r Runner) runBuild(cfg *config.Config, m *manifest.Manifest) error {
	switch m.Build.Engine {
	case "buildkit":
		if err := r.materializeContainerfile(m); err != nil {
			return err
		}
		imageRef := imageRefForProject(cfg, m)
		output := fmt.Sprintf("type=image,name=%s,push=false", imageRef)
		if err := build.RunTarget(build.Options{
			Workspace:  r.Workspace,
			Dockerfile: m.Build.Buildkit.Dockerfile,
			Target:     m.BuildkitTarget("image"),
			CacheRef:   m.Build.Buildkit.CacheRef,
			Output:     output,
		}); err != nil {
			return err
		}
		return outputs.Merge(r.Workspace, outputs.Entry{
			Name: "app",
			Type: "image",
			Ref:  imageRef,
		})
	case "buildpack":
		if m.Build.Buildpack == nil {
			return fmt.Errorf("manifest build.buildpack is required")
		}
		imageRef := imageRefForProject(cfg, m)
		output := fmt.Sprintf("type=image,name=%s,push=false", imageRef)
		if err := build.RunBuildpack(build.BuildpackOptions{
			Workspace: r.Workspace,
			Builder:   m.Build.Buildpack.Builder,
			RunImage:  m.Build.Buildpack.RunImage,
			CacheRef:  m.Build.Buildpack.CacheRef,
			Output:    output,
			Env:       buildpackGoEnv(),
		}); err != nil {
			return err
		}
		return outputs.Merge(r.Workspace, outputs.Entry{
			Name: "app",
			Type: "image",
			Ref:  imageRef,
		})
	case "dockerfile":
		return r.runDockerfileEngineImage(cfg, m, false)
	default:
		return fmt.Errorf("unsupported build engine %q", m.Build.Engine)
	}
}

func (r Runner) runPublish(cfg *config.Config, m *manifest.Manifest) error {
	items, err := outputs.Load(r.Workspace)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Println("==> no outputs to publish")
		return nil
	}
	for _, item := range items {
		switch item.Type {
		case "image":
			if err := r.pushImageOutput(cfg, m, item); err != nil {
				return fmt.Errorf("publish image %s: %w", item.Name, err)
			}
		default:
			fmt.Printf("==> skip output %s type=%s (not implemented)\n", item.Name, item.Type)
		}
	}
	return nil
}

func (r Runner) pushImageOutput(cfg *config.Config, m *manifest.Manifest, item outputs.Entry) error {
	ref := strings.TrimSpace(item.Ref)
	if ref == "" {
		return fmt.Errorf("empty image ref for %s", item.Name)
	}
	switch m.Build.Engine {
	case "buildkit":
		if err := r.materializeContainerfile(m); err != nil {
			return err
		}
		output := fmt.Sprintf("type=image,name=%s,push=true", ref)
		if err := build.RunTarget(build.Options{
			Workspace:  r.Workspace,
			Dockerfile: m.Build.Buildkit.Dockerfile,
			Target:     m.BuildkitTarget("image"),
			CacheRef:   m.Build.Buildkit.CacheRef,
			Output:     output,
		}); err != nil {
			return err
		}
		fmt.Printf("==> published image %s (%s)\n", item.Name, ref)
		return nil
	case "buildpack":
		if m.Build.Buildpack == nil {
			return fmt.Errorf("manifest build.buildpack is required")
		}
		output := fmt.Sprintf("type=image,name=%s,push=true", ref)
		if err := build.RunBuildpack(build.BuildpackOptions{
			Workspace: r.Workspace,
			Builder:   m.Build.Buildpack.Builder,
			RunImage:  m.Build.Buildpack.RunImage,
			CacheRef:  m.Build.Buildpack.CacheRef,
			Output:    output,
			Env:       buildpackGoEnv(),
		}); err != nil {
			return err
		}
		fmt.Printf("==> published image %s (%s)\n", item.Name, ref)
		return nil
	case "dockerfile":
		if err := r.runDockerfileEngineImage(cfg, m, true); err != nil {
			return err
		}
		fmt.Printf("==> published image %s (%s)\n", item.Name, ref)
		return nil
	default:
		return publish.PushImage(ref)
	}
}

func imageRefForProject(cfg *config.Config, m *manifest.Manifest) string {
	registry := strings.TrimSpace(os.Getenv("COIN_REGISTRY_PREFIX"))
	if registry == "" {
		registry = "localhost:8082/coin-docker"
	}
	tag := m.GoldenPath.Version
	if override := strings.TrimSpace(os.Getenv("COIN_IMAGE_TAG")); override != "" {
		tag = override
	}
	return fmt.Sprintf("%s/%s:%s", strings.TrimRight(registry, "/"), cfg.Project.Name, tag)
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
