package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"coin.local/coin-api/internal/canary"
	"coin.local/coin-api/internal/catalog"
	"coin.local/coin-api/internal/pin"
	"coin.local/coin-api/internal/resolve"
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
	Audience        string `json:"audience,omitempty"`
	Line            string `json:"line,omitempty"`
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

	res, err := s.resolve.Resolve(ctx, name, "="+release.Version, resolve.ResolveOptions{})
	if err != nil {
		return PublishGPReleaseResult{}, fmt.Errorf("resolve after promote: %w", err)
	}

	hash, _ := res.Document["manifestHash"].(string)
	if err := s.resolve.RefreshWildcards(ctx, name, release.Version, res.Document, hash); err != nil {
		return PublishGPReleaseResult{}, fmt.Errorf("refresh pointers: %w", err)
	}

	url, _ := s.store.ManifestURL(ctx, name, release.Version)
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

	var pointers []PointerStatus
	if catalogRow.Latest != "" {
		hash, _ := s.store.ManifestHash(ctx, gpName, catalogRow.Latest)
		pointers = append(pointers, PointerStatus{
			Pin:             "*",
			Audience:        "stable",
			Line:            "stable",
			ResolvedVersion: catalogRow.Latest,
			ManifestHash:    hash,
		})
	}
	if catalogRow.LatestCanary != "" {
		hash, _ := s.store.ManifestHash(ctx, gpName, catalogRow.LatestCanary)
		pointers = append(pointers, PointerStatus{
			Pin:             "*",
			Audience:        "canary",
			Line:            "canary",
			ResolvedVersion: catalogRow.LatestCanary,
			ManifestHash:    hash,
		})
	} else if catalogRow.Latest != "" {
		pointers = append(pointers, PointerStatus{
			Pin:      "*",
			Audience: "canary",
			Line:     "canary",
		})
	}

	if catalogRow.Latest != "" {
		for _, pinRaw := range []string{
			"=" + catalogRow.Latest,
			pin.TildePointer(catalogRow.Latest),
			pin.CaretPointer(catalogRow.Latest),
		} {
			p, err := pin.Parse(pinRaw)
			if err != nil {
				continue
			}
			version, err := p.SelectBest(published, catalogRow.Latest)
			if err != nil {
				pointers = append(pointers, PointerStatus{Pin: pinRaw, Audience: "all"})
				continue
			}
			hash, _ := s.store.ManifestHash(ctx, gpName, version)
			pointers = append(pointers, PointerStatus{
				Pin:             pinRaw,
				Audience:        "all",
				ResolvedVersion: version,
				ManifestHash:    hash,
			})
		}
	}

	if pointers == nil {
		pointers = []PointerStatus{}
	}
	return CatalogOverview{Catalog: catalogRow, Pointers: pointers}, nil
}

func (s *Service) PolicyCheck(ctx context.Context, gpName, version string) (catalog.Policy, string, error) {
	policy, err := s.store.GetCatalogPolicy(ctx, gpName)
	if err != nil {
		return catalog.Policy{}, "", err
	}
	warning, err := catalog.CheckValidate(policy, version)
	return policy, warning, err
}

func (s *Service) UpdateCatalogPolicy(ctx context.Context, gpName, latest, latestCanary, minimum string, deprecated []string, actor string) error {
	if err := s.store.ValidateCatalogPolicyUpdate(ctx, gpName, latest, latestCanary, minimum, deprecated); err != nil {
		return err
	}
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

func (s *Service) ListBuildReports(ctx context.Context, f store.ListBuildReportsFilter) ([]store.BuildReportRow, error) {
	return s.store.ListBuildReports(ctx, f)
}

func (s *Service) ListProjects(ctx context.Context, goldenPath, version string, staleOnly bool) ([]store.ProjectRow, error) {
	return s.store.ListProjects(ctx, goldenPath, version, staleOnly)
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

func (s *Service) CreateComponent(ctx context.Context, typ, name, actor string) error {
	return s.store.CreateComponent(ctx, typ, name, actor)
}

func (s *Service) ListComponentVersions(ctx context.Context, typ, name string) ([]store.ComponentVersionListItem, error) {
	return s.store.ListComponentVersions(ctx, typ, name)
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

func (s *Service) CreateGPProfileByAgentStack(ctx context.Context, name, agentStack, actor string) error {
	return s.store.CreateGPProfileByAgentStackWithDefaults(ctx, name, agentStack, actor)
}

func (s *Service) SaveComponentArtifact(ctx context.Context, typ, name, version, key string, body []byte) error {
	sum := sha256.Sum256(body)
	hash := "sha256:" + hex.EncodeToString(sum[:])
	return s.store.SaveComponentArtifactBody(ctx, typ, name, version, key, body, hash)
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
	CanaryContext   *CanaryContext `json:"canaryContext,omitempty"`
}

type ResolvePreviewOptions struct {
	Project      string
	ForceChannel string
}

type CanaryContext struct {
	Project        string `json:"project"`
	GPName         string `json:"gpName"`
	CanaryMode     string `json:"canaryMode"`
	RolloutEnabled bool   `json:"rolloutEnabled"`
	CanaryPercent  int    `json:"canaryPercent"`
	ProjectBucket  int    `json:"projectBucket"`
	UseCanaryLine  bool   `json:"useCanaryLine"`
	StableVersion  string `json:"stableVersion"`
	CanaryVersion  string `json:"canaryVersion"`
}

func (s *Service) GetCanaryContext(ctx context.Context, gpName, project string) (CanaryContext, error) {
	if project == "" {
		return CanaryContext{}, fmt.Errorf("project required")
	}
	cpol, err := s.store.GetCanaryPolicy(ctx, gpName)
	if err != nil {
		return CanaryContext{}, err
	}
	catalogRow, err := s.store.GetCatalogPolicyRow(ctx, gpName)
	if err != nil {
		return CanaryContext{}, err
	}
	projectMode, err := s.store.GetProjectCanaryMode(ctx, project)
	if err != nil {
		return CanaryContext{}, err
	}
	useCanary := canary.UseCanaryLine(project, projectMode, cpol.CanaryPercent, cpol.Enabled)
	return CanaryContext{
		Project:        project,
		GPName:         gpName,
		CanaryMode:     projectMode,
		RolloutEnabled: cpol.Enabled,
		CanaryPercent:  cpol.CanaryPercent,
		ProjectBucket:  canary.ProjectBucket(project),
		UseCanaryLine:  useCanary,
		StableVersion:  catalogRow.Latest,
		CanaryVersion:  catalogRow.LatestCanary,
	}, nil
}

func (s *Service) ResolvePreview(ctx context.Context, name, pinRaw string, opts ResolvePreviewOptions) (ResolvePreviewResult, error) {
	res, err := s.resolve.Resolve(ctx, name, pinRaw, resolve.ResolveOptions{
		Project:      opts.Project,
		ForceChannel: opts.ForceChannel,
	})
	if err != nil {
		return ResolvePreviewResult{}, err
	}
	out := ResolvePreviewResult{
		Manifest:        res.Document,
		RequestedPin:    res.RequestedPin,
		ResolvedVersion: res.ResolvedVersion,
		Channel:         res.Channel,
	}
	if opts.Project != "" {
		ctxRow, err := s.GetCanaryContext(ctx, name, opts.Project)
		if err == nil {
			out.CanaryContext = &ctxRow
		}
	}
	return out, nil
}

func (s *Service) GetPlatformSettings(ctx context.Context) (store.PlatformSettings, error) {
	return s.store.GetPlatformSettings(ctx)
}

func (s *Service) UpdatePlatformSettings(ctx context.Context, row store.PlatformSettings, actor string) error {
	return s.store.UpdatePlatformSettings(ctx, row, actor)
}

func (s *Service) GetComponentDetail(ctx context.Context, typ, name string) (store.ComponentDetail, error) {
	return s.store.GetComponentDetail(ctx, typ, name)
}

func (s *Service) GetComponentVersionDetail(ctx context.Context, typ, name, version string) (store.ComponentVersionDetail, error) {
	return s.store.GetComponentVersionDetail(ctx, typ, name, version)
}

type NextAgentVersionResult struct {
	Stack       string `json:"stack"`
	Runtime     string `json:"runtime"`
	CurrentRev  int    `json:"currentRev"`
	NextRev     int    `json:"nextRev"`
	NextVersion string `json:"nextVersion"`
}

func (s *Service) NextAgentVersion(ctx context.Context, stack, runtime string) (NextAgentVersionResult, error) {
	current, next, version, err := s.store.NextAgentVersion(ctx, stack, runtime)
	if err != nil {
		return NextAgentVersionResult{}, err
	}
	return NextAgentVersionResult{
		Stack:       stack,
		Runtime:     runtime,
		CurrentRev:  current,
		NextRev:     next,
		NextVersion: version,
	}, nil
}

type NextGPContentVersionResult struct {
	Name           string `json:"name"`
	Bump           string `json:"bump"`
	CurrentVersion string `json:"currentVersion"`
	NextVersion    string `json:"nextVersion"`
	IsFirst        bool   `json:"isFirst"`
}

func (s *Service) NextGPContentVersion(ctx context.Context, name, bump string) (NextGPContentVersionResult, error) {
	current, next, isFirst, err := s.store.NextGPContentVersion(ctx, name, bump)
	if err != nil {
		return NextGPContentVersionResult{}, err
	}
	return NextGPContentVersionResult{
		Name:           name,
		Bump:           bump,
		CurrentVersion: current,
		NextVersion:    next,
		IsFirst:        isFirst,
	}, nil
}

type NextLibVersionResult struct {
	Name           string `json:"name"`
	Bump           string `json:"bump"`
	CurrentVersion string `json:"currentVersion"`
	NextVersion    string `json:"nextVersion"`
	IsFirst        bool   `json:"isFirst"`
}

func (s *Service) NextLibVersion(ctx context.Context, name, bump string) (NextLibVersionResult, error) {
	current, next, isFirst, err := s.store.NextLibVersion(ctx, name, bump)
	if err != nil {
		return NextLibVersionResult{}, err
	}
	return NextLibVersionResult{
		Name:           name,
		Bump:           bump,
		CurrentVersion: current,
		NextVersion:    next,
		IsFirst:        isFirst,
	}, nil
}

type NextExecutorVersionResult struct {
	Name           string `json:"name"`
	Bump           string `json:"bump"`
	CurrentVersion string `json:"currentVersion"`
	NextVersion    string `json:"nextVersion"`
	IsFirst        bool   `json:"isFirst"`
}

func (s *Service) NextExecutorVersion(ctx context.Context, name, bump string) (NextExecutorVersionResult, error) {
	current, next, isFirst, err := s.store.NextExecutorVersion(ctx, name, bump)
	if err != nil {
		return NextExecutorVersionResult{}, err
	}
	return NextExecutorVersionResult{
		Name:           name,
		Bump:           bump,
		CurrentVersion: current,
		NextVersion:    next,
		IsFirst:        isFirst,
	}, nil
}
