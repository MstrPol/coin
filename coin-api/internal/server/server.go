package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"coin.local/coin-api/internal/admin"
	"coin.local/coin-api/internal/auth"
	"coin.local/coin-api/internal/catalog"
	"coin.local/coin-api/internal/config"
	"coin.local/coin-api/internal/metrics"
	"coin.local/coin-api/internal/nexus"
	"coin.local/coin-api/internal/report"
	"coin.local/coin-api/internal/resolve"
	"coin.local/coin-api/internal/store"
)

type Server struct {
	cfg     config.Config
	pool    *pgxpool.Pool
	logger  *slog.Logger
	resolve *resolve.Service
	report  *report.Service
	admin   *admin.Service
	oidc    *auth.OIDCVerifier
}

func New(cfg config.Config, pool *pgxpool.Pool, logger *slog.Logger) *Server {
	st := store.New(pool)
	nx := nexus.NewFromEnv()
	resolveSvc := resolve.New(st, nx)
	return &Server{
		cfg:     cfg,
		pool:    pool,
		logger:  logger,
		resolve: resolveSvc,
		report:  report.New(pool),
		admin:   admin.New(st, resolveSvc, logger),
		oidc:    auth.NewOIDCVerifier(cfg),
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	shortTimeout := middleware.Timeout(60 * time.Second)
	scanTimeout := middleware.Timeout(2 * time.Hour)

	r.Get("/health", s.health)
	r.Get("/ready", s.ready)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.Bearer(s.cfg))
		r.With(shortTimeout).Get("/golden-paths/{name}/versions/{version}/manifest", s.resolveManifest)
		r.With(shortTimeout).Get("/golden-paths/{name}/resolve", s.resolveByPin)
		r.With(shortTimeout).Post("/builds/report", s.buildReport)

		r.Route("/admin", func(r chi.Router) {
			r.Use(auth.AdminAuth(s.cfg, s.oidc))
			r.With(shortTimeout).Get("/me", s.adminMe)

			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(auth.RoleReader))
				r.With(shortTimeout).Get("/stats", s.adminStats)
				r.With(shortTimeout).Get("/projects", s.listProjects)
				r.With(shortTimeout).Get("/golden-paths", s.listGPReleases)
				r.With(shortTimeout).Get("/golden-paths/names", s.listGPNames)
				r.With(shortTimeout).Get("/golden-paths/{name}/profile", s.getGPProfile)
				r.With(shortTimeout).Get("/golden-paths/{name}/versions/{version}", s.getGPRelease)
				r.With(shortTimeout).Get("/golden-paths/{name}/catalog", s.getCatalog)
				r.With(shortTimeout).Get("/golden-paths/{name}/canary", s.getCanary)
				r.With(shortTimeout).Get("/golden-paths/{name}/versions/{version}/health", s.getHealth)
				r.With(shortTimeout).Get("/golden-paths/{name}/versions/{version}/artifacts", s.listArtifacts)
				r.With(shortTimeout).Get("/golden-paths/{name}/versions/{version}/artifacts/{key}", s.getArtifact)
				r.With(shortTimeout).Get("/golden-paths/{name}/resolve-preview", s.resolvePreview)
				r.With(shortTimeout).Get("/components", s.listComponents)
				r.With(shortTimeout).Get("/components/{type}/{name}/versions", s.listComponentVersions)
				r.With(shortTimeout).Get("/audit-log", s.listAuditLog)
				r.With(shortTimeout).Get("/golden-paths/{name}/versions/{version}/blast-radius", s.blastRadius)
			})

			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(auth.RolePublisher))
				r.With(shortTimeout).Post("/components/{type}/{name}/versions", s.publishComponentVersion)
				r.With(shortTimeout).Post("/golden-paths/{name}/versions", s.publishGPRelease)
				r.With(shortTimeout).Post("/golden-paths/{name}/drafts", s.createDraftGPRelease)
				r.With(shortTimeout).Post("/golden-paths/{name}/versions/{version}/promote", s.promoteDraftGPRelease)
				r.With(shortTimeout).Patch("/golden-paths/{name}/catalog", s.updateCatalog)
				r.With(shortTimeout).Patch("/golden-paths/{name}/canary", s.updateCanary)
				r.With(shortTimeout).Patch("/projects/{name}/canary-mode", s.updateProjectCanaryMode)
				r.With(shortTimeout).Post("/golden-paths/profiles", s.createGPProfile)
				r.With(shortTimeout).Put("/golden-paths/{name}/versions/{version}/artifacts/{key}", s.putArtifact)
			})

			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(auth.RoleAdmin))
				r.With(scanTimeout).Post("/scan", s.runFleetScan)
			})
		})
	})

	return r
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.pool.Ping(ctx); err != nil {
		s.logger.Warn("ready check failed", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
			"reason": "database",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) resolveByPin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	pinRaw := r.URL.Query().Get("pin")
	if pinRaw == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pin query parameter is required"})
		return
	}

	start := time.Now()
	status := "ok"
	defer func() {
		metrics.ObserveResolve(name, pinRaw, status, time.Since(start))
	}()

	res, err := s.resolve.Resolve(r.Context(), name, pinRaw, resolve.ResolveOptions{
		Project: r.URL.Query().Get("project"),
	})
	if errors.Is(err, store.ErrNotFound) {
		status = "not_found"
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "gp release not found"})
		return
	}
	if errors.Is(err, catalog.ErrBelowMinimum) {
		status = "forbidden"
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		status = "error"
		s.logger.Error("resolve by pin", "err", err, "gp", name, "pin", pinRaw)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if res.Warning != "" {
		w.Header().Set("Warning", res.Warning)
	}
	w.Header().Set("X-Coin-Requested-Pin", res.RequestedPin)
	w.Header().Set("X-Coin-Resolved-Version", res.ResolvedVersion)
	w.Header().Set("X-Coin-Channel", res.Channel)
	writeJSON(w, http.StatusOK, res.Document)
}

func (s *Server) resolveManifest(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	start := time.Now()
	status := "ok"

	defer func() {
		metrics.ObserveResolve(name, version, status, time.Since(start))
	}()

	doc, warning, err := s.resolve.Manifest(r.Context(), name, version)
	if errors.Is(err, store.ErrNotFound) {
		status = "not_found"
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "gp release not found"})
		return
	}
	if errors.Is(err, catalog.ErrBelowMinimum) {
		status = "forbidden"
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		status = "error"
		s.logger.Error("resolve manifest", "err", err, "gp", name, "version", version)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "resolve failed"})
		return
	}
	if warning != "" {
		w.Header().Set("Warning", warning)
	}
	writeJSON(w, http.StatusOK, doc)
}

type buildReportBody struct {
	Project         string `json:"project"`
	GoldenPath      string `json:"goldenPath"`
	Version         string `json:"version"`
	Branch          string `json:"branch"`
	BuildURL        string `json:"buildUrl"`
	Result          string `json:"result"`
	ManifestHash    string `json:"manifestHash"`
	GitURL          string `json:"gitUrl"`
	Channel         string `json:"channel"`
	RequestedPin    string `json:"requestedPin"`
	FailedStage     string `json:"failedStage"`
	ResolvedVersion string `json:"resolvedVersion"`
}

func (s *Server) buildReport(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req buildReportBody
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	id, err := s.report.Save(r.Context(), report.Input{
		Project:         req.Project,
		GoldenPath:      req.GoldenPath,
		Version:         req.Version,
		Branch:          req.Branch,
		BuildURL:        req.BuildURL,
		Result:          req.Result,
		ManifestHash:    req.ManifestHash,
		GitURL:          req.GitURL,
		Channel:         req.Channel,
		RequestedPin:    req.RequestedPin,
		FailedStage:     req.FailedStage,
		ResolvedVersion: req.ResolvedVersion,
	})
	if err != nil {
		s.logger.Error("build report", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "status": "accepted"})
}

type publishComponentBody struct {
	Version    string          `json:"version"`
	Metadata   json.RawMessage `json:"metadata"`
	ContentRef json.RawMessage `json:"contentRef"`
	Actor      string          `json:"actor"`
}

func (s *Server) publishComponentVersion(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req publishComponentBody
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Version == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "version is required"})
		return
	}

	row, err := s.admin.PublishComponentVersion(r.Context(), chi.URLParam(r, "type"), chi.URLParam(r, "name"), admin.PublishComponentRequest{
		Version:    req.Version,
		Metadata:   req.Metadata,
		ContentRef: req.ContentRef,
		Actor:      req.Actor,
	})
	if errors.Is(err, store.ErrDuplicateVersion) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "component version already exists"})
		return
	}
	if err != nil {
		s.logger.Error("publish component version", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":          row.ID,
		"componentId": row.ComponentID,
		"type":        row.Type,
		"name":        row.Name,
		"version":     row.Version,
		"status":      row.Status,
	})
}

type publishGPReleaseBody struct {
	Version     string            `json:"version"`
	Composition map[string]string `json:"composition"`
	Actor       string            `json:"actor"`
}

type createDraftBody struct {
	Version     string            `json:"version"`
	Composition map[string]string `json:"composition"`
	Actor       string            `json:"actor"`
}

func (s *Server) publishGPRelease(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req publishGPReleaseBody
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Version == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "version is required"})
		return
	}

	result, err := s.admin.PublishGPRelease(r.Context(), chi.URLParam(r, "name"), admin.PublishGPReleaseRequest{
		Version:     req.Version,
		Composition: req.Composition,
		Actor:       req.Actor,
	})
	if errors.Is(err, store.ErrDuplicateGPRelease) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "gp release already exists"})
		return
	}
	if errors.Is(err, store.ErrInvalidComposition) || errors.Is(err, store.ErrComponentNotFound) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		s.logger.Error("publish gp release", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "publish failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":              result.Release.ID,
		"name":            result.Release.Name,
		"version":         result.Release.Version,
		"status":          result.Release.Status,
		"manifestHash":    result.ManifestHash,
		"manifestUrl":     result.ManifestURL,
		"resolvedVersion": result.ResolvedVersion,
	})
}

func (s *Server) createDraftGPRelease(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req createDraftBody
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Version == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "version is required"})
		return
	}

	release, err := s.admin.CreateDraftGPRelease(r.Context(), chi.URLParam(r, "name"), admin.CreateDraftRequest{
		Version:     req.Version,
		Composition: req.Composition,
		Actor:       req.Actor,
	})
	if errors.Is(err, store.ErrDuplicateGPRelease) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "gp release already exists"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":      release.ID,
		"name":    release.Name,
		"version": release.Version,
		"status":  release.Status,
	})
}

func (s *Server) promoteDraftGPRelease(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	actor := r.URL.Query().Get("actor")

	result, err := s.admin.PromoteDraftGPRelease(r.Context(), name, version, actor)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "draft not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":              result.Release.ID,
		"name":            result.Release.Name,
		"version":         result.Release.Version,
		"status":          result.Release.Status,
		"manifestHash":    result.ManifestHash,
		"manifestUrl":     result.ManifestURL,
		"resolvedVersion": result.ResolvedVersion,
	})
}

func (s *Server) getCatalog(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	overview, err := s.admin.GetCatalogOverview(r.Context(), name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, overview)
}

type updateCatalogBody struct {
	Latest       string   `json:"latest"`
	LatestCanary string   `json:"latestCanary"`
	Minimum      string   `json:"minimum"`
	Deprecated   []string `json:"deprecated"`
	Actor        string   `json:"actor"`
}

func (s *Server) updateCatalog(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req updateCatalogBody
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := s.admin.UpdateCatalogPolicy(r.Context(), name, req.Latest, req.LatestCanary, req.Minimum, req.Deprecated, req.Actor); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) listArtifacts(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	items, err := s.admin.ListArtifactMeta(r.Context(), name, version)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if items == nil {
		items = []store.ArtifactMeta{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) getArtifact(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	key := chi.URLParam(r, "key")
	artifact, err := s.admin.GetArtifact(r.Context(), name, version, key)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "artifact not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"key":    artifact.Key,
		"sha256": artifact.SHA256,
		"body":   string(artifact.Body),
	})
}

func (s *Server) putArtifact(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	key := chi.URLParam(r, "key")
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := s.admin.SaveArtifact(r.Context(), name, version, key, []byte(req.Body)); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (s *Server) resolvePreview(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	pinRaw := r.URL.Query().Get("pin")
	if pinRaw == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pin required"})
		return
	}
	res, err := s.admin.ResolvePreview(r.Context(), name, pinRaw, r.URL.Query().Get("project"))
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"requestedPin":    res.RequestedPin,
		"resolvedVersion": res.ResolvedVersion,
		"channel":         res.Channel,
		"manifest":        res.Manifest,
	})
}

func (s *Server) blastRadius(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	result, err := s.admin.BlastRadius(r.Context(), name, version)
	if err != nil {
		s.logger.Error("blast radius", "err", err, "gp", name, "version", version)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "blast radius failed"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) adminMe(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}
	roles := make([]string, len(p.Roles))
	for i, role := range p.Roles {
		roles[i] = string(role)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"subject":    p.Subject,
		"email":      p.Email,
		"roles":      roles,
		"authMethod": p.AuthMethod,
	})
}

func (s *Server) adminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.admin.DashboardStats(r.Context())
	if err != nil {
		s.logger.Error("admin stats", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "stats failed"})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, err := s.admin.ListProjects(r.Context(), q.Get("goldenPath"), q.Get("version"))
	if err != nil {
		s.logger.Error("list projects", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list projects failed"})
		return
	}
	if rows == nil {
		rows = []store.ProjectRow{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": rows})
}

func (s *Server) listGPReleases(w http.ResponseWriter, r *http.Request) {
	includeDrafts := r.URL.Query().Get("includeDrafts") == "true"
	var rows []store.GPReleaseListItem
	var err error
	if includeDrafts {
		rows, err = s.admin.ListGPReleasesAll(r.Context(), r.URL.Query().Get("name"))
	} else {
		rows, err = s.admin.ListGPReleases(r.Context(), r.URL.Query().Get("name"))
	}
	if err != nil {
		s.logger.Error("list gp releases", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list gp releases failed"})
		return
	}
	if rows == nil {
		rows = []store.GPReleaseListItem{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": rows})
}

func (s *Server) getGPRelease(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	detail, err := s.admin.GetGPRelease(r.Context(), name, version)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "gp release not found"})
		return
	}
	if err != nil {
		s.logger.Error("get gp release", "err", err, "gp", name, "version", version)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "get gp release failed"})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) listComponents(w http.ResponseWriter, r *http.Request) {
	rows, err := s.admin.ListComponents(r.Context())
	if err != nil {
		s.logger.Error("list components", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list components failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": rows})
}

func (s *Server) listComponentVersions(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	rows, err := s.admin.ListComponentVersions(r.Context(), typ, name)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "component not found"})
		return
	}
	if err != nil {
		s.logger.Error("list component versions", "err", err, "type", typ, "name", name)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list component versions failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": rows})
}

func (s *Server) runFleetScan(w http.ResponseWriter, r *http.Request) {
	force := r.URL.Query().Get("force") == "true"
	result, err := s.admin.RunFleetScan(r.Context(), force)
	if err != nil {
		s.logger.Error("fleet scan", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":  "scan failed",
			"result": result,
		})
		return
	}
	status := http.StatusOK
	if result.ReposFailed > 0 {
		status = http.StatusMultiStatus
	}
	writeJSON(w, status, map[string]any{
		"status": "complete",
		"result": result,
	})
}

func (s *Server) listGPNames(w http.ResponseWriter, r *http.Request) {
	names, err := s.admin.ListGPNames(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": names})
}

func (s *Server) getGPProfile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	profile, err := s.admin.GetGPProfile(r.Context(), name)
	if errors.Is(err, store.ErrUnsupportedGP) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "golden path not found"})
		return
	}
	if err != nil {
		s.logger.Error("gp profile", "err", err, "gp", name)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "gp profile failed"})
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) listAuditLog(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 50
	offset := 0
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}
	rows, err := s.admin.ListAuditLog(r.Context(), store.AuditLogFilter{
		EntityType: q.Get("entityType"),
		Action:     q.Get("action"),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		s.logger.Error("audit log", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "audit log failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": rows})
}

func (s *Server) getCanary(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	overview, err := s.admin.GetCanaryOverview(r.Context(), name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) updateCanary(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req struct {
		Enabled                     bool   `json:"enabled"`
		CanaryPercent               int    `json:"canaryPercent"`
		DegradedThresholdPct        int    `json:"degradedThresholdPct"`
		CriticalThresholdPct        int    `json:"criticalThresholdPct"`
		CriticalConsecutiveFailures int    `json:"criticalConsecutiveFailures"`
		Actor                       string `json:"actor"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	row := store.CanaryPolicyRow{
		GPName:                      name,
		Enabled:                     req.Enabled,
		CanaryPercent:               req.CanaryPercent,
		DegradedThresholdPct:        req.DegradedThresholdPct,
		CriticalThresholdPct:        req.CriticalThresholdPct,
		CriticalConsecutiveFailures: req.CriticalConsecutiveFailures,
	}
	if err := s.admin.UpdateCanaryPolicy(r.Context(), row, req.Actor); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) updateProjectCanaryMode(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req struct {
		Mode  string `json:"mode"`
		Actor string `json:"actor"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := s.admin.SetProjectCanaryMode(r.Context(), projectName, req.Mode, req.Actor); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) getHealth(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	channel := r.URL.Query().Get("channel")
	summary, err := s.admin.GetHealth(r.Context(), name, version, channel)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) createGPProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req struct {
		Name  string                `json:"name"`
		Slots []store.GPProfileSlot `json:"slots"`
		Actor string                `json:"actor"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := s.admin.CreateGPProfile(r.Context(), req.Name, req.Slots, req.Actor); err != nil {
		if errors.Is(err, store.ErrDuplicateGPProfile) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "gp profile already exists"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "name": req.Name})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
