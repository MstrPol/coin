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
	"coin.local/coin-executor/pkg/branching"
)

type RunOptions struct {
	Stage string
}

type Runner struct {
	Workspace string
}

func (r Runner) Run(cfg *config.Config, m *manifest.Manifest, opts RunOptions) error {
	if err := m.ValidateDeliverables(); err != nil {
		return err
	}
	items := m.DeliverableSpecs()
	if err := deliverables.Validate(items, deliverables.P0Types); err != nil {
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
		if key == "publish" {
			if err := enforcePublishPolicy(r.Workspace, m); err != nil {
				return err
			}
		} else if opts.Stage == "" && !shouldRunStage(stage) {
			fmt.Printf("==> skip stage %s (when=%s)\n", key, stage.When)
			continue
		}
		fmt.Printf("==> stage %s\n", key)
		if len(stage.Steps) > 0 {
			if err := r.runStageSteps(cfg, m, stage); err != nil {
				return fmt.Errorf("stage %s: %w", key, err)
			}
			continue
		}
		if err := r.runStage(cfg, m, key); err != nil {
			return fmt.Errorf("stage %s: %w", key, err)
		}
	}
	return nil
}

func (r Runner) runStage(cfg *config.Config, m *manifest.Manifest, stageID string) error {
	switch stageID {
	case "validate":
		return validate.Project(cfg, m, r.Workspace)
	case "test":
		return r.runTest(cfg, m)
	case "build":
		return r.runBuild(cfg, m)
	case "publish":
		return r.runPublish(cfg, m)
	default:
		return fmt.Errorf("unknown stage %q", stageID)
	}
}

func (r Runner) runStageSteps(cfg *config.Config, m *manifest.Manifest, stage manifest.Stage) error {
	for _, step := range stage.Steps {
		switch step.Action {
		case "run":
			if err := r.runInlineRun(cfg, m, step.Run); err != nil {
				return err
			}
		case "build":
			if err := r.runInlineBuild(cfg, m, step.Build, false); err != nil {
				return err
			}
		case "publish":
			if err := r.runInlinePublish(cfg, m, step.Publish); err != nil {
				return err
			}
		case "run-target":
			if err := r.runVNextTarget(cfg, m, step.TargetID, ""); err != nil {
				return err
			}
		case "build-deliverable":
			if err := r.buildVNextDeliverable(cfg, m, step.DeliverableID, false); err != nil {
				return err
			}
		case "publish-deliverable":
			if err := r.buildVNextDeliverable(cfg, m, step.DeliverableID, true); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown stage step action %q", step.Action)
		}
	}
	return nil
}

func (r Runner) runInlineRun(cfg *config.Config, m *manifest.Manifest, run *manifest.InlineRunStep) error {
	if run == nil {
		return fmt.Errorf("run step block is required")
	}
	output := ""
	if run.Output != "" {
		output = run.Output
	}
	target := strings.TrimSpace(run.Target)
	if target == "" {
		target = inferStageTarget("run", output, "")
	}
	return r.runInlineEngine(cfg, m, run.Engine, target, run.Containerfile, run.Dockerfile, output)
}

func (r Runner) runInlineBuild(cfg *config.Config, m *manifest.Manifest, build *manifest.InlineBuildStep, push bool) error {
	if build == nil {
		return fmt.Errorf("build step block is required")
	}
	ref := ""
	output := ""
	switch build.Type {
	case "image", "liquibase-image":
		item := manifest.Deliverable{ID: build.ID, Type: build.Type, Image: build.Image}
		ref = imageRefForDeliverable(cfg, m, item, r.Workspace)
		output = fmt.Sprintf("type=image,name=%s,push=%t", ref, push)
	case "artifact":
		fmt.Printf("==> build artifact deliverable %s\n", build.ID)
	default:
		return fmt.Errorf("unsupported build type %q", build.Type)
	}
	if err := r.runInlineEngine(cfg, m, build.Engine, inferBuildTarget(build), build.Containerfile, build.Dockerfile, output); err != nil {
		return err
	}
	if push {
		return nil
	}
	entry := outputs.Entry{Name: build.ID, Type: build.Type, Ref: ref}
	if build.Type == "artifact" {
		entry.Format = build.Artifact.Format
		if entry.Format == "" {
			entry.Format = "zip"
		}
		entry.Ref = artifactRepositoryURL(m)
	}
	return outputs.Merge(r.Workspace, entry)
}

func inferBuildTarget(build *manifest.InlineBuildStep) string {
	if build == nil {
		return ""
	}
	if target := strings.TrimSpace(build.Target); target != "" {
		return target
	}
	return inferStageTarget("build", "", build.Type)
}

func (r Runner) runInlinePublish(cfg *config.Config, m *manifest.Manifest, publish *manifest.InlinePublishStep) error {
	if publish == nil || strings.TrimSpace(publish.BuildStepID) == "" {
		return fmt.Errorf("publish.buildStepId is required")
	}
	build, ok := m.InlineBuild(publish.BuildStepID)
	if !ok {
		return fmt.Errorf("build step %q not found", publish.BuildStepID)
	}
	cp := build
	return r.runInlineBuild(cfg, m, &cp, true)
}

func (r Runner) runInlineEngine(cfg *config.Config, m *manifest.Manifest, engine, target string, cf *manifest.InlineContainerfileStep, df *manifest.InlineDockerfileStep, output string) error {
	switch engine {
	case "buildkit":
		dockerfile, err := r.materializeStepContainerfile(m, cf)
		if err != nil {
			return err
		}
		return build.RunTarget(build.Options{
			Workspace:  r.Workspace,
			Dockerfile: dockerfile,
			Target:     target,
			CacheRef:   cacheRefForProject(cfg, m),
			Output:     output,
		})
	case "dockerfile":
		path := ""
		if df != nil {
			path = df.Path
		}
		tgt := target
		if tgt == "" && df != nil {
			tgt = df.Target
		}
		if tgt == "" {
			tgt = inferStageTarget("run", output, "")
		}
		return build.RunTarget(build.Options{
			Workspace:  r.Workspace,
			Dockerfile: path,
			Target:     tgt,
			CacheRef:   cacheRefForProject(cfg, m),
			Output:     output,
		})
	default:
		return fmt.Errorf("unsupported engine %q", engine)
	}
}

func (r Runner) buildVNextDeliverable(cfg *config.Config, m *manifest.Manifest, deliverableID string, push bool) error {
	item, ok := m.Deliverable(deliverableID)
	if !ok {
		return fmt.Errorf("deliverable %q not found", deliverableID)
	}
	ref := ""
	output := ""
	switch item.Type {
	case "image", "liquibase-image":
		ref = imageRefForDeliverable(cfg, m, item, r.Workspace)
		output = fmt.Sprintf("type=image,name=%s,push=%t", ref, push)
	case "artifact":
		fmt.Printf("==> build artifact deliverable %s\n", item.ID)
	default:
		return fmt.Errorf("unsupported deliverable type %q", item.Type)
	}
	if err := r.runVNextTarget(cfg, m, item.TargetID, output); err != nil {
		return err
	}
	if push {
		return nil
	}
	entry := outputs.Entry{Name: item.ID, Type: item.Type, Ref: ref}
	if item.Type == "artifact" {
		entry.Format = item.Artifact.Format
		if entry.Format == "" {
			entry.Format = "zip"
		}
		entry.Ref = artifactRepositoryURL(m)
	}
	return outputs.Merge(r.Workspace, entry)
}

func (r Runner) runVNextTarget(cfg *config.Config, m *manifest.Manifest, targetID string, output string) error {
	target, ok := m.Target(targetID)
	if !ok {
		return fmt.Errorf("target %q not found", targetID)
	}
	switch target.Engine {
	case "buildkit":
		dockerfile, err := r.materializeNamedContainerfile(m, target.Containerfile)
		if err != nil {
			return err
		}
		return build.RunTarget(build.Options{
			Workspace:  r.Workspace,
			Dockerfile: dockerfile,
			Target:     target.Target,
			CacheRef:   cacheRefForProject(cfg, m),
			Output:     output,
		})
	case "dockerfile":
		return build.RunTarget(build.Options{
			Workspace:  r.Workspace,
			Dockerfile: target.Dockerfile,
			Target:     target.Target,
			CacheRef:   cacheRefForProject(cfg, m),
			Output:     output,
		})
	default:
		return fmt.Errorf("unsupported target engine %q", target.Engine)
	}
}

func (r Runner) runTest(cfg *config.Config, m *manifest.Manifest) error {
	switch m.Build.Engine {
	case "buildkit":
		return r.runBuildkitTarget(cfg, m, m.BuildkitTarget("test"))
	case "dockerfile":
		target := ""
		if m.Build.Dockerfile != nil {
			target = m.Build.Dockerfile.TestTarget
		}
		if target == "" {
			fmt.Println("==> skip test (dockerfile engine: testTarget not set)")
			return nil
		}
		return r.runDockerfileEngineTarget(cfg, m, target)
	default:
		return fmt.Errorf("unsupported build engine %q for test stage", m.Build.Engine)
	}
}

func (r Runner) runBuildkitTarget(cfg *config.Config, m *manifest.Manifest, target string) error {
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
		CacheRef:   cacheRefForProject(cfg, m),
	})
}

func (r Runner) runBuild(cfg *config.Config, m *manifest.Manifest) error {
	switch m.Build.Engine {
	case "buildkit":
		if !m.HasDeliverable("image") {
			fmt.Println("==> skip image build (GP manifest has no image deliverable)")
			return nil
		}
		if err := r.materializeContainerfile(m); err != nil {
			return err
		}
		imageRef := imageRefForProject(cfg, m, r.Workspace)
		output := fmt.Sprintf("type=image,name=%s,push=false", imageRef)
		if err := build.RunTarget(build.Options{
			Workspace:  r.Workspace,
			Dockerfile: m.Build.Buildkit.Dockerfile,
			Target:     m.BuildkitTarget("image"),
			CacheRef:   cacheRefForProject(cfg, m),
			Output:     output,
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
			CacheRef:   cacheRefForProject(cfg, m),
			Output:     output,
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

func imageRefForProject(cfg *config.Config, m *manifest.Manifest, workDir string) string {
	tag := strings.TrimSpace(os.Getenv("COIN_IMAGE_TAG"))
	if tag == "" {
		tag = strings.TrimSpace(os.Getenv("COIN_VERSION"))
	}
	if tag == "" {
		if model := branching.FromManifest(m); model != nil {
			if g, err := branching.GitFromEnv(workDir); err == nil {
				if v, err := branching.ResolveVersion(model, g); err == nil {
					tag = v
				}
			}
		}
	}
	if tag == "" {
		tag = m.GoldenPath.Version
	}
	return fmt.Sprintf("%s:%s", imageRepositoryForProject(cfg, m, ""), tag)
}

func imageRefForDeliverable(cfg *config.Config, m *manifest.Manifest, item manifest.Deliverable, workDir string) string {
	suffix := strings.TrimSpace(item.Image.RepositorySuffix)
	if item.Type == "liquibase-image" && suffix == "" {
		suffix = "-liquibase"
	}
	tag := strings.TrimSpace(os.Getenv("COIN_IMAGE_TAG"))
	if tag == "" {
		tag = strings.TrimSpace(os.Getenv("COIN_VERSION"))
	}
	if tag == "" {
		if model := branching.FromManifest(m); model != nil {
			if g, err := branching.GitFromEnv(workDir); err == nil {
				if v, err := branching.ResolveVersion(model, g); err == nil {
					tag = v
				}
			}
		}
	}
	if tag == "" {
		tag = m.GoldenPath.Version
	}
	return fmt.Sprintf("%s:%s", imageRepositoryForProject(cfg, m, suffix), tag)
}

func cacheRefForProject(cfg *config.Config, m *manifest.Manifest) string {
	if !m.Destinations.BuildCacheEnabled {
		return ""
	}
	return imageRepositoryForProject(cfg, m, "-cache")
}

func artifactRepositoryURL(m *manifest.Manifest) string {
	return strings.TrimRight(strings.TrimSpace(m.Destinations.ArtifactRepositoryBase), "/")
}

func imageRepositoryForProject(cfg *config.Config, m *manifest.Manifest, suffix string) string {
	registryPrefix := strings.TrimRight(strings.TrimSpace(m.Destinations.ImageRegistryPrefix), "/")
	if runtimePrefix := strings.TrimRight(strings.TrimSpace(os.Getenv("COIN_REGISTRY_PREFIX")), "/"); runtimePrefix != "" {
		registryPrefix = runtimePrefix
	}
	return fmt.Sprintf(
		"%s/%s/%s/%s%s",
		registryPrefix,
		strings.Trim(strings.TrimSpace(cfg.Project.GroupID), "/"),
		strings.Trim(strings.TrimSpace(cfg.Project.ArtifactID), "/"),
		strings.Trim(strings.TrimSpace(cfg.Project.Name), "/"),
		suffix,
	)
}

func enforcePublishPolicy(workDir string, m *manifest.Manifest) error {
	model := branching.FromManifest(m)
	if model == nil {
		return nil
	}
	g, err := branching.GitFromEnv(workDir)
	if err != nil {
		return fmt.Errorf("branching git: %w", err)
	}
	return branching.CheckPublishAllowed(model, g)
}

func shouldSkipPublish(workDir string, m *manifest.Manifest) (bool, string) {
	// Legacy: without branching section, honor pipeline when=tag.
	if branching.FromManifest(m) != nil {
		return false, ""
	}
	// Legacy manifests without branching: pipeline stage when=tag.
	for _, stage := range m.Pipeline.Stages {
		if stage.Key() == "publish" && !shouldRunStage(stage) {
			return true, fmt.Sprintf("pipeline when=%s", stage.When)
		}
	}
	return false, ""
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
