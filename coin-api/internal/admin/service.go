package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"coin.local/coin-api/internal/pin"
	"coin.local/coin-api/internal/resolve"
	"coin.local/coin-api/internal/scanner"
	"coin.local/coin-api/internal/store"
)

type Service struct {
	store   *store.Store
	resolve *resolve.Service
	logger  *slog.Logger
}

func New(st *store.Store, rs *resolve.Service, logger *slog.Logger) *Service {
	return &Service{store: st, resolve: rs, logger: logger}
}

type PublishComponentRequest struct {
	Version    string          `json:"version"`
	Metadata   json.RawMessage `json:"metadata"`
	ContentRef json.RawMessage `json:"contentRef"`
	Actor      string          `json:"actor"`
}

type PublishGPReleaseRequest struct {
	Version     string            `json:"version"`
	Composition map[string]string `json:"composition"`
	Actor       string            `json:"actor"`
}

type CreateDraftRequest struct {
	Version     string            `json:"version"`
	Composition map[string]string `json:"composition"`
	Actor       string            `json:"actor"`
}

type PublishGPReleaseResult struct {
	Release         store.GPReleaseRow
	ManifestHash    string
	ManifestURL     string
	ResolvedVersion string
}

type PointerStatus struct {
	Pin             string `json:"pin"`
	ResolvedVersion string `json:"resolvedVersion"`
	ManifestHash    string `json:"manifestHash,omitempty"`
}

type CatalogOverview struct {
	Catalog  store.CatalogPolicyRow `json:"catalog"`
	Pointers []PointerStatus        `json:"pointers"`
}

func (s *Service) PublishComponentVersion(ctx context.Context, typ, name string, req PublishComponentRequest) (store.ComponentVersionRow, error) {
	return s.store.PublishComponentVersion(ctx, store.ComponentVersionInput{
		Type:       typ,
		Name:       name,
		Version:    req.Version,
		Metadata:   req.Metadata,
		ContentRef: req.ContentRef,
		Actor:      req.Actor,
	})
}

func (s *Service) CreateDraftGPRelease(ctx context.Context, name string, req CreateDraftRequest) (store.GPReleaseRow, error) {
	return s.store.CreateDraftGPRelease(ctx, store.PublishGPReleaseInput{
		Name:        name,
		Version:     req.Version,
		Composition: req.Composition,
		Actor:       req.Actor,
	})
}

func (s *Service) PublishGPRelease(ctx context.Context, name string, req PublishGPReleaseRequest) (PublishGPReleaseResult, error) {
	release, err := s.store.PublishGPRelease(ctx, store.PublishGPReleaseInput{
		Name:        name,
		Version:     req.Version,
		Composition: req.Composition,
		Actor:       req.Actor,
	})
	if err != nil {
		return PublishGPReleaseResult{}, err
	}

	res, err := s.resolve.Resolve(ctx, name, "="+req.Version, resolve.ResolveOptions{})
	if err != nil {
		return PublishGPReleaseResult{}, fmt.Errorf("resolve after publish: %w", err)
	}

	hash, _ := res.Document["manifestHash"].(string)
	if err := s.resolve.RefreshWildcards(ctx, name, req.Version, res.Document, hash); err != nil {
		return PublishGPReleaseResult{}, fmt.Errorf("refresh pointers: %w", err)
	}

	url, _ := s.store.ManifestURL(ctx, name, req.Version)

	return PublishGPReleaseResult{
		Release:         release,
		ManifestHash:    hash,
		ManifestURL:     url,
		ResolvedVersion: res.ResolvedVersion,
	}, nil
}

func (s *Service) PromoteDraftGPRelease(ctx context.Context, name, version, actor string) (PublishGPReleaseResult, error) {
	release, err := s.store.PromoteDraftToPublished(ctx, name, version, actor)
	if err != nil {
		return PublishGPReleaseResult{}, err
	}

	res, err := s.resolve.Resolve(ctx, name, "="+version, resolve.ResolveOptions{})
	if err != nil {
		return PublishGPReleaseResult{}, fmt.Errorf("resolve after promote: %w", err)
	}

	hash, _ := res.Document["manifestHash"].(string)
	if err := s.resolve.RefreshWildcards(ctx, name, version, res.Document, hash); err != nil {
		return PublishGPReleaseResult{}, fmt.Errorf("refresh pointers: %w", err)
	}

	url, _ := s.store.ManifestURL(ctx, name, version)
	return PublishGPReleaseResult{
		Release:         release,
		ManifestHash:    hash,
		ManifestURL:     url,
		ResolvedVersion: res.ResolvedVersion,
	}, nil
}

func (s *Service) GetCatalogOverview(ctx context.Context, gpName string) (CatalogOverview, error) {
	catalogRow, err := s.store.GetCatalogPolicyRow(ctx, gpName)
	if err != nil {
		return CatalogOverview{}, err
	}

	published, err := s.store.ListPublishedGPVersions(ctx, gpName)
	if err != nil {
		return CatalogOverview{}, err
	}

	pinKeys := []string{"*"}
	if catalogRow.Latest != "" {
		pinKeys = append(pinKeys,
			"="+catalogRow.Latest,
			pin.TildePointer(catalogRow.Latest),
			pin.CaretPointer(catalogRow.Latest),
		)
	}

	var pointers []PointerStatus
	for _, pinRaw := range pinKeys {
		p, err := pin.Parse(pinRaw)
		if err != nil {
			continue
		}
		version, err := p.SelectBest(published, catalogRow.Latest)
		if err != nil {
			pointers = append(pointers, PointerStatus{Pin: pinRaw})
			continue
		}
		hash, _ := s.store.ManifestHash(ctx, gpName, version)
		pointers = append(pointers, PointerStatus{
			Pin:             pinRaw,
			ResolvedVersion: version,
			ManifestHash:    hash,
		})
	}

	if catalogRow.LatestCanary != "" {
		hash, _ := s.store.ManifestHash(ctx, gpName, catalogRow.LatestCanary)
		pointers = append(pointers, PointerStatus{
			Pin:             "canary:latest",
			ResolvedVersion: catalogRow.LatestCanary,
			ManifestHash:    hash,
		})
	}

	return CatalogOverview{Catalog: catalogRow, Pointers: pointers}, nil
}

func (s *Service) UpdateCatalogPolicy(ctx context.Context, gpName, latest, latestCanary, minimum string, deprecated []string, actor string) error {
	return s.store.UpdateCatalogPolicy(ctx, gpName, latest, latestCanary, minimum, deprecated, actor)
}

func (s *Service) ListArtifactMeta(ctx context.Context, gpName, version string) ([]store.ArtifactMeta, error) {
	return s.store.ListArtifactMeta(ctx, gpName, version)
}

func (s *Service) GetArtifact(ctx context.Context, gpName, version, key string) (store.ArtifactBody, error) {
	return s.store.GetArtifactBody(ctx, gpName, version, key)
}

func (s *Service) SaveArtifact(ctx context.Context, gpName, version, key string, body []byte) error {
	return s.store.UpsertArtifactBody(ctx, gpName, version, key, body)
}

func (s *Service) BlastRadius(ctx context.Context, name, version string) (store.BlastRadius, error) {
	return s.store.BlastRadius(ctx, name, version)
}

func (s *Service) DashboardStats(ctx context.Context) (store.DashboardStats, error) {
	return s.store.DashboardStats(ctx)
}

func (s *Service) ListProjects(ctx context.Context, goldenPath, version string) ([]store.ProjectRow, error) {
	return s.store.ListProjects(ctx, goldenPath, version)
}

func (s *Service) ListGPReleases(ctx context.Context, name string) ([]store.GPReleaseListItem, error) {
	return s.store.ListGPReleases(ctx, name, false)
}

func (s *Service) ListGPReleasesAll(ctx context.Context, name string) ([]store.GPReleaseListItem, error) {
	return s.store.ListGPReleases(ctx, name, true)
}

func (s *Service) GetGPRelease(ctx context.Context, name, version string) (store.GPReleaseDetail, error) {
	return s.store.GetGPReleaseDetail(ctx, name, version)
}

func (s *Service) ListComponents(ctx context.Context) ([]store.ComponentListItem, error) {
	return s.store.ListComponents(ctx)
}

func (s *Service) ListComponentVersions(ctx context.Context, typ, name string) ([]store.ComponentVersionListItem, error) {
	return s.store.ListComponentVersions(ctx, typ, name)
}

func (s *Service) RunFleetScan(ctx context.Context, force bool) (store.ScanResult, error) {
	svc := scanner.New(scanner.NewGiteaFromEnv(), s.store, s.logger)
	return svc.Run(ctx, force)
}

func (s *Service) ListAuditLog(ctx context.Context, f store.AuditLogFilter) ([]store.AuditLogEntry, error) {
	return s.store.ListAuditLog(ctx, f)
}

func (s *Service) ListGPNames(ctx context.Context) ([]string, error) {
	return s.store.ListGPNames(ctx)
}

func (s *Service) GetGPProfile(ctx context.Context, name string) (store.GPProfile, error) {
	return s.store.GetGPProfile(ctx, name)
}

func (s *Service) CreateGPProfile(ctx context.Context, name string, slots []store.GPProfileSlot, actor string) error {
	return s.store.CreateGPProfileWithDefaults(ctx, name, slots, actor)
}

type CanaryOverview struct {
	Policy       store.CanaryPolicyRow `json:"policy"`
	Catalog      store.CatalogPolicyRow `json:"catalog"`
	InCanary     int                   `json:"inCanary"`
	TotalProjects int                  `json:"totalProjects"`
}

func (s *Service) GetCanaryOverview(ctx context.Context, gpName string) (CanaryOverview, error) {
	policy, err := s.store.GetCanaryPolicy(ctx, gpName)
	if err != nil {
		return CanaryOverview{}, err
	}
	catalog, err := s.store.GetCatalogPolicyRow(ctx, gpName)
	if err != nil {
		return CanaryOverview{}, err
	}
	inCanary, total, err := s.store.CountProjectsInCanaryBucket(ctx, policy.CanaryPercent)
	if err != nil {
		return CanaryOverview{}, err
	}
	return CanaryOverview{
		Policy:        policy,
		Catalog:       catalog,
		InCanary:      inCanary,
		TotalProjects: total,
	}, nil
}

func (s *Service) UpdateCanaryPolicy(ctx context.Context, row store.CanaryPolicyRow, actor string) error {
	return s.store.UpsertCanaryPolicy(ctx, row, actor)
}

func (s *Service) SetProjectCanaryMode(ctx context.Context, projectName, mode, actor string) error {
	return s.store.SetProjectCanaryMode(ctx, projectName, mode, actor)
}

func (s *Service) GetHealth(ctx context.Context, gpName, version, channel string) (store.HealthSummary, error) {
	policy, err := s.store.GetCanaryPolicy(ctx, gpName)
	if err != nil {
		return store.HealthSummary{}, err
	}
	if channel == "" {
		channel = "canary"
	}
	return s.store.AggregateHealth(ctx, gpName, version, channel, 24*time.Hour, policy)
}

type ResolvePreviewResult struct {
	Manifest        map[string]any
	RequestedPin    string
	ResolvedVersion string
	Channel         string
}

func (s *Service) ResolvePreview(ctx context.Context, name, pinRaw, project string) (ResolvePreviewResult, error) {
	res, err := s.resolve.Resolve(ctx, name, pinRaw, resolve.ResolveOptions{Project: project})
	if err != nil {
		return ResolvePreviewResult{}, err
	}
	return ResolvePreviewResult{
		Manifest:        res.Document,
		RequestedPin:    res.RequestedPin,
		ResolvedVersion: res.ResolvedVersion,
		Channel:         res.Channel,
	}, nil
}
