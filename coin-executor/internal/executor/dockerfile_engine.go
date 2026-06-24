package executor

import (
	"fmt"
	"path/filepath"

	"coin.local/coin-executor/internal/build"
	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/content"
	"coin.local/coin-executor/internal/manifest"
	"coin.local/coin-executor/internal/outputs"
)

func (r Runner) materializeContainerfile(m *manifest.Manifest) error {
	if m.Build.Buildkit == nil {
		return fmt.Errorf("manifest build.buildkit is required")
	}
	return r.materializeContentFile(
		m.Build.Buildkit.Dockerfile,
		m.Build.Buildkit.Containerfile,
		m,
	)
}

func (r Runner) materializeDockerfileEngine(m *manifest.Manifest) error {
	if m.Build.Dockerfile == nil {
		return fmt.Errorf("manifest build.dockerfile is required")
	}
	return r.materializeContentFile(
		m.Build.Dockerfile.Dockerfile,
		m.Build.Dockerfile.Containerfile,
		m,
	)
}

func (r Runner) materializeContentFile(destPath string, cref manifest.ContentRef, m *manifest.Manifest) error {
	dest := destPath
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(r.Workspace, dest)
	}
	contentRoot, err := content.EnsureRoot(m)
	if err != nil {
		return err
	}
	return content.Materialize(dest, contentRoot, cref)
}

func (r Runner) runDockerfileEngineTarget(m *manifest.Manifest, target string) error {
	if m.Build.Dockerfile == nil {
		return fmt.Errorf("manifest build.dockerfile is required")
	}
	if err := r.materializeDockerfileEngine(m); err != nil {
		return err
	}
	return build.RunTarget(build.Options{
		Workspace:  r.Workspace,
		Dockerfile: m.Build.Dockerfile.Dockerfile,
		Target:     target,
		CacheRef:   m.Build.Dockerfile.CacheRef,
	})
}

func (r Runner) runDockerfileEngineImage(cfg *config.Config, m *manifest.Manifest, push bool) error {
	if m.Build.Dockerfile == nil {
		return fmt.Errorf("manifest build.dockerfile is required")
	}
	if err := r.materializeDockerfileEngine(m); err != nil {
		return err
	}
	imageRef := imageRefForProject(cfg, m, r.Workspace)
	pushFlag := "false"
	if push {
		pushFlag = "true"
	}
	output := fmt.Sprintf("type=image,name=%s,push=%s", imageRef, pushFlag)
	if err := build.RunTarget(build.Options{
		Workspace:  r.Workspace,
		Dockerfile: m.Build.Dockerfile.Dockerfile,
		Target:     m.Build.Dockerfile.ImageTarget,
		CacheRef:   m.Build.Dockerfile.CacheRef,
		Output:     output,
	}); err != nil {
		return err
	}
	if !push {
		return outputs.Merge(r.Workspace, outputs.Entry{
			Name: "app",
			Type: "image",
			Ref:  imageRef,
		})
	}
	return nil
}
