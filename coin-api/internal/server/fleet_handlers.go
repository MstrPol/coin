package server

import (
	"encoding/csv"
	"net/http"
	"net/url"
	"strconv"

	"coin.local/coin-api/internal/store"
)

func parseProjectsFilter(q url.Values) store.ListProjectsFilter {
	staleOnly := q.Get("stale") == "true" || q.Get("stale") == "1"
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	return store.ListProjectsFilter{
		GoldenPath: q.Get("goldenPath"),
		Version:    q.Get("version"),
		StaleOnly:  staleOnly,
		Limit:      limit,
		Offset:     offset,
	}
}

func parseBuildReportsFilter(q url.Values) (store.ListBuildReportsFilter, error) {
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	after, err := store.ParseQueryDateStart(q.Get("reportedAfter"))
	if err != nil {
		return store.ListBuildReportsFilter{}, err
	}
	before, err := store.ParseQueryDateEnd(q.Get("reportedBefore"))
	if err != nil {
		return store.ListBuildReportsFilter{}, err
	}
	return store.ListBuildReportsFilter{
		Project:        q.Get("project"),
		GoldenPath:     q.Get("goldenPath"),
		Result:         q.Get("result"),
		ReportedAfter:  after,
		ReportedBefore: before,
		Limit:          limit,
		Offset:         offset,
	}, nil
}

func writePaginatedJSON[T any](w http.ResponseWriter, items []T, total, limit, offset int) {
	if items == nil {
		items = []T{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) listBuildReports(w http.ResponseWriter, r *http.Request) {
	f, err := parseBuildReportsFilter(r.URL.Query())
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	limit := f.Limit
	offset := f.Offset
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	f.Limit = limit
	f.Offset = offset

	total, err := s.admin.CountBuildReports(r.Context(), f)
	if err != nil {
		s.logger.Error("count build reports", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "count build reports failed"})
		return
	}
	items, err := s.admin.ListBuildReports(r.Context(), f)
	if err != nil {
		s.logger.Error("list build reports", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list build reports failed"})
		return
	}
	writePaginatedJSON(w, items, total, limit, offset)
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	f := parseProjectsFilter(r.URL.Query())
	limit := f.Limit
	offset := f.Offset
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	f.Limit = limit
	f.Offset = offset

	total, err := s.admin.CountProjects(r.Context(), f)
	if err != nil {
		s.logger.Error("count projects", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "count projects failed"})
		return
	}
	rows, err := s.admin.ListProjects(r.Context(), f)
	if err != nil {
		s.logger.Error("list projects", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list projects failed"})
		return
	}
	writePaginatedJSON(w, rows, total, limit, offset)
}

func (s *Server) exportProjects(w http.ResponseWriter, r *http.Request) {
	f := parseProjectsFilter(r.URL.Query())
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+store.FormatProjectExportFilename()+`"`)
	cw := csv.NewWriter(w)
	if err := s.admin.WriteProjectsCSV(r.Context(), f, cw); err != nil {
		s.logger.Error("export projects csv", "err", err)
		return
	}
	cw.Flush()
}

func (s *Server) exportBuildReports(w http.ResponseWriter, r *http.Request) {
	f, err := parseBuildReportsFilter(r.URL.Query())
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+store.FormatBuildReportsExportFilename()+`"`)
	cw := csv.NewWriter(w)
	if err := s.admin.WriteBuildReportsCSV(r.Context(), f, cw); err != nil {
		s.logger.Error("export build reports csv", "err", err)
		return
	}
	cw.Flush()
}
