package resolve

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"coin.local/coin-api/internal/canary"
	"coin.local/coin-api/internal/catalog"
	"coin.local/coin-api/internal/manifest"
	"coin.local/coin-api/internal/nexus"
	"coin.local/coin-api/internal/pin"
	"coin.local/coin-api/internal/store"
)

type Service struct {
	store   *store.Store
	builder manifest.Builder
	nexus   *nexus.Client
}

type Result struct {
	Document        map[string]any
	Warning         string
	RequestedPin    string
	ResolvedVersion string
	Channel         string // stable | canary
}

func New(st *store.Store, nx *nexus.Client) *Service {
	return &Service{store: st, builder: manifest.Builder{}, nexus: nx}
}

func (s *Service) Manifest(ctx context.Context, name, version string) (map[string]any, string, error) {
	res, err := s.Resolve(ctx, name, "="+version, ResolveOptions{})
	if err != nil {
		return nil, "", err
	}
	return res.Document, res.Warning, nil
}

type ResolveOptions struct {
	Project      string
	ForceChannel string // canary | stable — только resolve-preview
}

func (s *Service) Resolve(ctx context.Context, name, pinRaw string, opts ResolveOptions) (Result, error) {
	p, err := pin.Parse(pinRaw)
	if err != nil {
		return Result{}, fmt.Errorf("invalid pin: %w", err)
	}

	channel := "stable"
	version, err := s.selectVersion(ctx, name, p, opts, &channel)
	if err != nil {
		return Result{}, err
	}

	policy, err := s.store.GetCatalogPolicy(ctx, name)
	if err != nil {
		return Result{}, err
	}
	warning, err := catalog.CheckResolve(policy, version)
	if err != nil {
		return Result{}, err
	}

	allowDraft := p.Kind == pin.KindExact && pin.IsSnapshotVersion(p.Base)
	release, err := s.store.GetGPReleaseForResolve(ctx, name, version, allowDraft)
	if errors.Is(err, store.ErrNotFound) {
		return Result{}, err
	}
	if err != nil {
		return Result{}, err
	}

	doc, hash, err := s.builder.Build(manifest.GPRelease{
		Name:    release.Name,
		Version: release.Version,
		Parts:   release.Parts,
		Content: release.Content,
	}, manifest.BuildOptions{
		Project:      opts.Project,
		RegistryHost: registryHostForManifest(),
	})
	if err != nil {
		return Result{}, err
	}

	if s.nexus != nil {
		if err := s.publishToNexus(ctx, name, release.Version, p.Raw, doc, hash); err != nil {
			_ = err
		}
	}

	return Result{
		Document:        doc,
		Warning:         warning,
		RequestedPin:    p.Raw,
		ResolvedVersion: release.Version,
		Channel:         channel,
	}, nil
}

func (s *Service) selectVersion(ctx context.Context, name string, p pin.Pin, opts ResolveOptions, channel *string) (string, error) {
	if p.Kind == pin.KindExact {
		allowDraft := pin.IsSnapshotVersion(p.Base)
		_, err := s.store.GetGPReleaseForResolve(ctx, name, p.Base, allowDraft)
		if err != nil {
			return "", err
		}
		*channel = "stable"
		return p.Base, nil
	}

	published, err := s.store.ListPublishedGPVersions(ctx, name)
	if err != nil {
		return "", err
	}
	policy, err := s.store.GetCatalogPolicy(ctx, name)
	if err != nil {
		return "", err
	}

	if p.Kind == pin.KindLatest {
		cpol, err := s.store.GetCanaryPolicy(ctx, name)
		if err != nil {
			return "", err
		}
		projectMode, err := s.store.GetProjectCanaryMode(ctx, opts.Project)
		if err != nil {
			return "", err
		}
		useCanary := canary.UseCanaryLine(opts.Project, projectMode, cpol.CanaryPercent, cpol.Enabled)
		switch opts.ForceChannel {
		case "canary":
			useCanary = true
		case "stable":
			useCanary = false
		}
		if useCanary {
			if policy.LatestCanary == "" {
				return "", fmt.Errorf("canary line not configured for %s", name)
			}
			*channel = "canary"
			return policy.LatestCanary, nil
		}
		*channel = "stable"
		return p.SelectBest(published, policy.Latest)
	}

	*channel = "stable"
	return p.SelectBest(published, policy.Latest)
}

func (s *Service) publishToNexus(ctx context.Context, gpName, version, requestedPin string, doc map[string]any, hash string) error {
	blobURL, err := s.nexus.UploadManifestBlob(ctx, gpName, version, doc)
	if err != nil {
		return err
	}
	ptr := nexus.PointerDoc{
		ResolvedVersion: version,
		ManifestHash:    hash,
		BlobURL:         blobURL,
	}

	if _, err := s.nexus.PutPointer(ctx, gpName, requestedPin, ptr); err != nil {
		return err
	}
	exactPin := "=" + version
	if requestedPin != exactPin {
		if _, err := s.nexus.PutPointer(ctx, gpName, exactPin, ptr); err != nil {
			return err
		}
	}

	_ = s.store.SaveManifestMeta(ctx, gpName, version, hash, blobURL)
	s.publishContentArtifacts(ctx, gpName, version)
	return nil
}

func (s *Service) publishContentArtifacts(ctx context.Context, gpName, version string) {
	artifacts, err := s.store.ListArtifactBodies(ctx, gpName, version)
	if err != nil {
		return
	}
	for _, a := range artifacts {
		_, _ = s.nexus.UploadContentArtifact(ctx, gpName, version, a.Key, a.Body)
	}
}

// RefreshWildcards updates ~ / ^ / * pointers when a new published version becomes best match.
type LibraryVersionResult struct {
	GoldenPath      string `json:"goldenPath"`
	ResolvedVersion string `json:"resolvedVersion"`
	Library         struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"library"`
}

// LibraryVersion resolves coin-lib version for product Jenkins bootstrap.
func (s *Service) LibraryVersion(ctx context.Context, name, pinRaw string, opts ResolveOptions) (LibraryVersionResult, error) {
	p, err := pin.Parse(pinRaw)
	if err != nil {
		return LibraryVersionResult{}, fmt.Errorf("invalid pin: %w", err)
	}

	channel := "stable"
	version, err := s.selectVersion(ctx, name, p, opts, &channel)
	if err != nil {
		return LibraryVersionResult{}, err
	}

	libName, libVer, err := s.store.LibVersionFromComposition(ctx, name, version)
	if err != nil {
		return LibraryVersionResult{}, err
	}

	var out LibraryVersionResult
	out.GoldenPath = name
	out.ResolvedVersion = version
	out.Library.Type = "lib"
	out.Library.Name = libName
	out.Library.Version = libVer
	return out, nil
}

func (s *Service) RefreshWildcards(ctx context.Context, gpName, version string, doc map[string]any, hash string) error {
	if s.nexus == nil {
		return nil
	}
	published, err := s.store.ListPublishedGPVersions(ctx, gpName)
	if err != nil {
		return err
	}
	policy, err := s.store.GetCatalogPolicy(ctx, gpName)
	if err != nil {
		return err
	}

	blobURL, err := s.nexus.UploadManifestBlob(ctx, gpName, version, doc)
	if err != nil {
		return err
	}
	ptr := nexus.PointerDoc{
		ResolvedVersion: version,
		ManifestHash:    hash,
		BlobURL:         blobURL,
	}

	_, _ = s.nexus.PutPointer(ctx, gpName, "="+version, ptr)

	for _, pinKey := range []string{pin.TildePointer(version), pin.CaretPointer(version)} {
		p, _ := pin.Parse(pinKey)
		best, err := p.SelectBest(published, "")
		if err == nil && best == version {
			if _, err := s.nexus.PutPointer(ctx, gpName, pinKey, ptr); err != nil {
				return err
			}
		}
	}

	if policy.Latest == version {
		if _, err := s.nexus.PutPointer(ctx, gpName, "*", ptr); err != nil {
			return err
		}
	}
	return nil
}

func registryHostForManifest() string {
	if host := strings.TrimSpace(os.Getenv("COIN_REGISTRY_HOST")); host != "" {
		return strings.TrimSuffix(host, "/")
	}
	return "nexus:8082"
}
