package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"coin.local/coin-api/internal/admin"
	"coin.local/coin-api/internal/auth"
	"coin.local/coin-api/internal/catalog"
	"coin.local/coin-api/internal/config"
	"coin.local/coin-api/internal/docs"
	"coin.local/coin-api/internal/metrics"
	"coin.local/coin-api/internal/version"
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
		admin:   admin.New(st, resolveSvc, nx, logger),
		oidc:    auth.NewOIDCVerifier(cfg),
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	shortTimeout := middleware.Timeout(60 * time.Second)

	docHandler := docs.NewHandler()

	r.Get("/health", s.health)
	r.Get("/ready", s.ready)
	r.Get("/openapi/v1.yaml", docHandler.ServeOpenAPI)
	r.Get("/docs", docHandler.ServeSwaggerUI)
	r.Get("/docs/", docHandler.ServeSwaggerUI)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.Bearer(s.cfg))
		r.With(shortTimeout).Get("/golden-paths/{name}/versions/{version}/manifest", s.resolveManifest)
		r.With(shortTimeout).Get("/golden-paths/{name}/resolve", s.resolveByPin)
		r.With(shortTimeout).Get("/golden-paths/{name}/policy-check", s.policyCheck)
		r.With(shortTimeout).Post("/builds/report", s.buildReport)

		r.Route("/admin", func(r chi.Router) {
			r.Use(auth.AdminAuth(s.cfg, s.oidc))
			r.With(shortTimeout).Get("/me", s.adminMe)

			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(auth.RoleReader))
				r.With(shortTimeout).Get("/stats", s.adminStats)
				r.With(shortTimeout).Get("/build-reports", s.listBuildReports)
				r.With(shortTimeout).Get("/build-reports/export", s.exportBuildReports)
				r.With(shortTimeout).Get("/projects", s.listProjects)
				r.With(shortTimeout).Get("/projects/export", s.exportProjects)
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
				r.With(shortTimeout).Post("/branching-models/preview", s.branchingModelPreview)
				r.With(shortTimeout).Get("/golden-paths/{name}/projects/{project}/canary-context", s.canaryContext)
				r.With(shortTimeout).Get("/components", s.listComponents)
				r.With(shortTimeout).Get("/components/agent/{name}/next-version", s.nextAgentVersion)
				r.With(shortTimeout).Get("/components/gp-content/{name}/next-version", s.nextGPContentVersion)
				r.With(shortTimeout).Get("/components/executor/{name}/next-version", s.nextExecutorVersion)
				r.With(shortTimeout).Get("/components/{type}/{name}", s.getComponentDetail)
				r.With(shortTimeout).Get("/components/{type}/{name}/versions", s.listComponentVersions)
				r.With(shortTimeout).Get("/components/{type}/{name}/versions/{version}", s.getComponentVersionDetail)
				r.With(shortTimeout).Get("/audit-log", s.listAuditLog)
				r.With(shortTimeout).Get("/golden-paths/{name}/versions/{version}/blast-radius", s.blastRadius)
			})

			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(auth.RolePublisher))
				r.With(shortTimeout).Post("/components", s.createComponent)
				r.With(shortTimeout).Post("/components/{type}/{name}/versions", s.publishComponentVersion)
				r.With(shortTimeout).Post("/components/{type}/{name}/versions/drafts", s.createDraftComponentVersion)
				r.With(shortTimeout).Post("/components/{type}/{name}/versions/{version}/promote", s.promoteComponentVersion)
				r.With(shortTimeout).Patch("/components/{type}/{name}/versions/{version}", s.patchComponentVersion)
				r.With(shortTimeout).Post("/components/{type}/{name}/versions/{version}/register-package", s.registerComponentPackage)
				r.With(shortTimeout).Post("/components/{type}/{name}/versions/{version}/validate-package", s.validateComponentPackage)
				r.With(shortTimeout).Put("/components/{type}/{name}/versions/{version}/artifacts/*", s.putComponentArtifact)
				r.With(shortTimeout).Get("/components/{type}/{name}/versions/{version}/artifacts", s.listComponentArtifacts)
				r.With(shortTimeout).Get("/components/{type}/{name}/versions/{version}/artifacts/*", s.getComponentArtifact)
				r.With(shortTimeout).Post("/golden-paths/{name}/versions", s.publishGPRelease)
				r.With(shortTimeout).Post("/golden-paths/{name}/drafts", s.createDraftGPRelease)
				r.With(shortTimeout).Post("/golden-paths/{name}/versions/{version}/promote", s.promoteDraftGPRelease)
				r.With(shortTimeout).Delete("/golden-paths/{name}/versions/{version}", s.deleteGPReleaseDraft)
				r.With(shortTimeout).Patch("/golden-paths/{name}/versions/{version}", s.updateGPReleaseDraft)
				r.With(shortTimeout).Patch("/golden-paths/{name}/catalog", s.updateCatalog)
				r.With(shortTimeout).Patch("/golden-paths/{name}/canary", s.updateCanary)
				r.With(shortTimeout).Patch("/projects/{name}/canary-mode", s.updateProjectCanaryMode)
				r.With(shortTimeout).Post("/golden-paths/profiles", s.createGPProfile)
				r.With(shortTimeout).Put("/golden-paths/{name}/versions/{version}/artifacts/{key}", s.putArtifact)
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
		writeJSON(w, http.StatusServiceUnavailable, readyPayload("not ready", "database"))
		return
	}
	writeJSON(w, http.StatusOK, readyPayload("ready", ""))
}

func readyPayload(status, reason string) map[string]string {
	body := map[string]string{
		"status":  status,
		"version": version.Version,
	}
	if reason != "" {
		body["reason"] = reason
	}
	return body
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
	GroupID         string `json:"groupId"`
	ArtifactID      string `json:"artifactId"`
	GoldenPath      string `json:"goldenPath"`
	Version         string `json:"version"`
	ConfigVersion   string `json:"configVersion"`
	Branch          string `json:"branch"`
	BuildURL        string `json:"buildUrl"`
	Result          string `json:"result"`
	ManifestHash    string `json:"manifestHash"`
	GitURL          string `json:"gitUrl"`
	Channel         string `json:"channel"`
	RequestedPin    string `json:"requestedPin"`
	FailedStage     string `json:"failedStage"`
	ResolvedVersion string           `json:"resolvedVersion"`
	Outputs         []map[string]any `json:"outputs"`
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
		GroupID:         req.GroupID,
		ArtifactID:      req.ArtifactID,
		GoldenPath:      req.GoldenPath,
		Version:         req.Version,
		ConfigVersion:   req.ConfigVersion,
		Branch:          req.Branch,
		BuildURL:        req.BuildURL,
		Result:          req.Result,
		ManifestHash:    req.ManifestHash,
		GitURL:          req.GitURL,
		Channel:         req.Channel,
		RequestedPin:    req.RequestedPin,
		FailedStage:     req.FailedStage,
		ResolvedVersion: req.ResolvedVersion,
		Outputs:         req.Outputs,
	})
	if err != nil {
		s.logger.Error("build report", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "status": "accepted"})
}

type createComponentBody struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Actor string `json:"actor"`
}

func (s *Server) createComponent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req createComponentBody
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Type == "" || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type and name are required"})
		return
	}
	if err := s.admin.CreateComponent(r.Context(), req.Type, req.Name, req.Actor); err != nil {
		if errors.Is(err, store.ErrDuplicateComponent) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "component already exists"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "type": req.Type, "name": req.Name})
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

func (s *Server) createDraftComponentVersion(w http.ResponseWriter, r *http.Request) {
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

	row, err := s.admin.CreateDraftComponentVersion(r.Context(), chi.URLParam(r, "type"), chi.URLParam(r, "name"), admin.PublishComponentRequest{
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
		s.logger.Error("create draft component version", "err", err)
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


func (s *Server) promoteComponentVersion(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	actor := actorFromBody(r)

	row, err := s.admin.PromoteComponentToPublished(r.Context(), typ, name, version, actor)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "component version not found"})
		return
	}
	if errors.Is(err, store.ErrComponentVersionNotDraft) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	if errors.Is(err, store.ErrComponentVersionNotCanary) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		s.logger.Error("promote component version", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"type":    row.Type,
		"name":    row.Name,
		"version": row.Version,
		"status":  row.Status,
	})
}

func actorFromBody(r *http.Request) string {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		return ""
	}
	var req struct {
		Actor string `json:"actor"`
	}
	if json.Unmarshal(body, &req) != nil {
		return ""
	}
	return req.Actor
}

func (s *Server) patchComponentVersion(w http.ResponseWriter, r *http.Request) {
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
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	if err := s.admin.UpdateComponentVersion(r.Context(), typ, name, version, admin.PublishComponentRequest{
		Metadata:   req.Metadata,
		ContentRef: req.ContentRef,
		Actor:      req.Actor,
	}); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "component version not found"})
			return
		}
		if errors.Is(err, store.ErrComponentVersionNotDraft) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		s.logger.Error("patch component version", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated", "type": typ, "name": name, "version": version})
}

func (s *Server) registerComponentPackage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req admin.RegisterComponentPackageRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
	}
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	result, err := s.admin.RegisterComponentPackage(r.Context(), typ, name, version, req)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "component version not found"})
			return
		}
		if errors.Is(err, store.ErrComponentVersionNotDraft) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, store.ErrInvalidContentRef) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		s.logger.Error("register component package", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) validateComponentPackage(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	result, err := s.admin.ValidateComponentPackage(r.Context(), typ, name, version)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "component version not found"})
			return
		}
		if errors.Is(err, store.ErrComponentVersionNotDraft) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		s.logger.Error("validate component package", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	status := http.StatusOK
	if !result.Valid {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, result)
}

type publishGPReleaseBody struct {
	Version            string            `json:"version"`
	Composition        map[string]string `json:"composition"`
	AgentStackName     string            `json:"agentStackName"`
	GPContentName      string            `json:"gpContentName"`
	BranchingModelName string            `json:"branchingModelName"`
	Actor              string            `json:"actor"`
}

type createDraftBody struct {
	Version            string            `json:"version"`
	Composition        map[string]string `json:"composition"`
	AgentStackName     string            `json:"agentStackName"`
	GPContentName      string            `json:"gpContentName"`
	BranchingModelName string            `json:"branchingModelName"`
	Actor              string            `json:"actor"`
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
		Version:            req.Version,
		Composition:        req.Composition,
		AgentStackName:     req.AgentStackName,
		GPContentName:      req.GPContentName,
		BranchingModelName: req.BranchingModelName,
		Actor:              req.Actor,
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
		Version:            req.Version,
		Composition:        req.Composition,
		AgentStackName:     req.AgentStackName,
		GPContentName:      req.GPContentName,
		BranchingModelName: req.BranchingModelName,
		Actor:              req.Actor,
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
	if errors.Is(err, store.ErrGPCompositionHasDraftPins) {
		blockers, _ := s.admin.ValidateGPReleasePromoteBlockers(r.Context(), name, version)
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":         err.Error(),
			"blockingPins": blockers,
		})
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

func (s *Server) deleteGPReleaseDraft(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	actor := r.URL.Query().Get("actor")

	err := s.admin.DeleteGPReleaseDraft(r.Context(), name, version, actor)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "gp release not found"})
		return
	}
	if errors.Is(err, store.ErrGPReleaseNotDraft) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "published releases are immutable"})
		return
	}
	if err != nil {
		s.logger.Error("delete gp draft", "err", err, "gp", name, "version", version)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete failed"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateGPReleaseDraft(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
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

	release, err := s.admin.UpdateGPReleaseDraft(r.Context(), name, version, admin.CreateDraftRequest{
		Composition:        req.Composition,
		AgentStackName:     req.AgentStackName,
		GPContentName:      req.GPContentName,
		BranchingModelName: req.BranchingModelName,
		Actor:              req.Actor,
	})
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "gp release not found"})
		return
	}
	if errors.Is(err, store.ErrGPReleaseNotDraft) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "published releases are immutable"})
		return
	}
	if errors.Is(err, store.ErrInvalidComposition) || errors.Is(err, store.ErrComponentNotFound) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		s.logger.Error("update gp draft", "err", err, "gp", name, "version", version)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    release.Name,
		"version": release.Version,
		"status":  release.Status,
	})
}

func (s *Server) policyCheck(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := r.URL.Query().Get("version")
	if version == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "version query parameter is required"})
		return
	}
	_, warning, err := s.admin.PolicyCheck(r.Context(), name, version)
	if errors.Is(err, catalog.ErrBelowMinimum) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"gpName":  name,
		"version": version,
		"warning": warning,
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
	writeListJSON(w, items)
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
	q := r.URL.Query()
	res, err := s.admin.ResolvePreview(r.Context(), name, pinRaw, admin.ResolvePreviewOptions{
		Project:      q.Get("project"),
		ForceChannel: q.Get("forceChannel"),
	})
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	out := map[string]any{
		"requestedPin":    res.RequestedPin,
		"resolvedVersion": res.ResolvedVersion,
		"channel":         res.Channel,
		"manifest":        res.Manifest,
	}
	if res.CanaryContext != nil {
		out["canaryContext"] = res.CanaryContext
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) canaryContext(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	project := chi.URLParam(r, "project")
	ctx, err := s.admin.GetCanaryContext(r.Context(), name, project)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ctx)
}

func (s *Server) getComponentDetail(w http.ResponseWriter, r *http.Request) {
	detail, err := s.admin.GetComponentDetail(r.Context(), chi.URLParam(r, "type"), chi.URLParam(r, "name"))
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "get component failed"})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) getComponentVersionDetail(w http.ResponseWriter, r *http.Request) {
	detail, err := s.admin.GetComponentVersionDetail(
		r.Context(),
		chi.URLParam(r, "type"),
		chi.URLParam(r, "name"),
		chi.URLParam(r, "version"),
	)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "get component version failed"})
		return
	}
	resp := map[string]any{
		"type":       detail.Type,
		"name":       detail.Name,
		"version":    detail.Version,
		"status":     detail.Status,
		"metadata":   json.RawMessage(detail.Metadata),
		"createdAt":  detail.CreatedAt,
	}
	if len(detail.ContentRef) > 0 {
		resp["contentRef"] = json.RawMessage(detail.ContentRef)
	}
	if detail.Type == "agent" {
		if pin, ok := store.DerivedExecutorPin(detail.Name, detail.Version); ok {
			resp["derivedExecutorPin"] = pin
		}
	}
	writeJSON(w, http.StatusOK, resp)
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
	writeListJSON(w, rows)
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
	writeListJSON(w, rows)
}

func (s *Server) nextAgentVersion(w http.ResponseWriter, r *http.Request) {
	stack := chi.URLParam(r, "name")
	runtime := r.URL.Query().Get("runtime")
	if runtime == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "runtime query param required"})
		return
	}
	res, err := s.admin.NextAgentVersion(r.Context(), stack, runtime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) nextGPContentVersion(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	bump := r.URL.Query().Get("bump")
	if bump == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bump query param required (major, minor, patch)"})
		return
	}
	res, err := s.admin.NextGPContentVersion(r.Context(), name, bump)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) nextExecutorVersion(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	bump := r.URL.Query().Get("bump")
	if bump == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bump query param required (major, minor, patch)"})
		return
	}
	res, err := s.admin.NextExecutorVersion(r.Context(), name, bump)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
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
	writeListJSON(w, rows)
}

func (s *Server) listGPNames(w http.ResponseWriter, r *http.Request) {
	names, err := s.admin.ListGPNames(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeListJSON(w, names)
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
	writeListJSON(w, rows)
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

func (s *Server) putComponentArtifact(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if decoded, err := url.PathUnescape(key); err == nil {
		key = decoded
	}
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req struct {
		Body       string `json:"body"`
		BodyBase64 string `json:"bodyBase64"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	var raw []byte
	switch {
	case req.Body != "":
		raw = []byte(req.Body)
	case req.BodyBase64 != "":
		raw, err = base64.StdEncoding.DecodeString(req.BodyBase64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bodyBase64"})
			return
		}
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body or bodyBase64 required"})
		return
	}
	if err := s.admin.SaveComponentArtifact(r.Context(), typ, name, version, key, raw); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (s *Server) listComponentArtifacts(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	items, err := s.admin.ListComponentArtifacts(r.Context(), typ, name, version)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "component version not found"})
			return
		}
		if errors.Is(err, store.ErrComponentVersionNotDraft) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeListJSON(w, items)
}

func (s *Server) getComponentArtifact(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if decoded, err := url.PathUnescape(key); err == nil {
		key = decoded
	}
	body, sha, err := s.admin.GetComponentArtifact(r.Context(), typ, name, version, key)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "artifact not found"})
			return
		}
		if errors.Is(err, store.ErrComponentVersionNotDraft) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"key":    key,
		"sha256": sha,
		"body":   string(body),
	})
}

func (s *Server) createGPProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Actor       string `json:"actor"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	createErr := s.admin.CreateGPProfile(r.Context(), req.Name, req.Description, req.Actor)
	if createErr != nil {
		if errors.Is(createErr, store.ErrDuplicateGPProfile) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "gp profile already exists"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": createErr.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "name": req.Name})
}

func writeListJSON[T any](w http.ResponseWriter, items []T) {
	if items == nil {
		items = []T{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
