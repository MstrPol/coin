package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"coin.local/coin-executor/internal/build"
	"coin.local/coin-executor/internal/config"
	"coin.local/coin-executor/internal/content"
	"coin.local/coin-executor/internal/manifest"
	"coin.local/coin-executor/internal/outputs"
)

const managedContainerfileDir = ".coin/containerfiles"

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

func (r Runner) materializeStepContainerfile(m *manifest.Manifest, cf *manifest.InlineContainerfileStep) (string, error) {
	if cf == nil {
		return "", fmt.Errorf("step containerfile is required")
	}
	destPath := filepath.Join(managedContainerfileDir, "step.Containerfile")
	cref := manifest.ContentRef{}
	if cf.ContentRef != nil {
		cref.URL = cf.ContentRef["url"]
		cref.SHA256 = cf.ContentRef["sha256"]
		if cref.SHA256 == "" {
			cref.SHA256 = cf.Digest
		}
	}
	if strings.TrimSpace(cref.URL) == "" && strings.TrimSpace(cf.Body) != "" {
		dest := filepath.Join(r.Workspace, destPath)
		if err := os.WriteFile(dest, []byte(cf.Body), 0o644); err != nil {
			return "", err
		}
		return destPath, nil
	}
	if strings.TrimSpace(cref.URL) == "" {
		return "", fmt.Errorf("step containerfile contentRef or body is required")
	}
	return destPath, r.materializeContentFile(destPath, cref, m)
}

func (r Runner) materializeNamedContainerfile(m *manifest.Manifest, id string) (string, error) {
	ref, ok := m.Containerfile(id)
	if !ok {
		return "", fmt.Errorf("manifest artifacts.containerfiles missing %q", id)
	}
	destPath := filepath.Join(managedContainerfileDir, id+".Containerfile")
	return destPath, r.materializeContentFile(destPath, manifest.ContentRef{
		URL:    ref.URL,
		SHA256: ref.SHA256,
	}, m)
}

func (r Runner) materializeDockerfileEngine(m *manifest.Manifest) error {
	if m.Build.Dockerfile == nil {
		return fmt.Errorf("manifest build.dockerfile is required")
	}
	if strings.TrimSpace(m.Build.Dockerfile.Containerfile.URL) == "" {
		return nil
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

func (r Runner) runDockerfileEngineTarget(cfg *config.Config, m *manifest.Manifest, target string) error {
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
		CacheRef:   cacheRefForProject(cfg, m),
	})
}

func (r Runner) runDockerfileEngineImage(cfg *config.Config, m *manifest.Manifest, push bool) error {
	if m.Build.Dockerfile == nil {
		return fmt.Errorf("manifest build.dockerfile is required")
	}
	if !m.HasDeliverable("image") {
		fmt.Println("==> skip image build (GP manifest has no image deliverable)")
		return nil
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
		CacheRef:   cacheRefForProject(cfg, m),
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
