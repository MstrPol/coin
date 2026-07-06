package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"coin.local/coin-api/internal/gpcontent"
	"coin.local/coin-api/internal/store"
)

type gpPipelinePreviewRequest struct {
	Model any `json:"model"`
}

func (s *Server) gpReleasePipelinePreview(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	var req gpPipelinePreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	if req.Model == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "model is required"})
		return
	}
	modelRaw, err := json.Marshal(req.Model)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid model"})
		return
	}
	raw, err := normalizeGPReleasePipelineModel(modelRaw, name)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	doc, err := gpcontent.ParseDoc(raw)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	doc.Name = name
	doc.Version = version
	result := gpcontent.Preview(doc, gpcontent.PreviewOptions{ComponentName: name})
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) getGPReleasePipeline(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	body, err := s.admin.GetGPReleasePipelineBody(r.Context(), name, version)
	if errors.Is(err, store.ErrNotFound) || errors.Is(err, store.ErrGPReleasePipelineMissing) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "pipeline body not found"})
		return
	}
	if err != nil {
		s.logger.Error("get gp release pipeline", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "load failed"})
		return
	}
	writeJSON(w, http.StatusOK, body)
}

func (s *Server) putGPReleasePipeline(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	defer r.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	raw, err := normalizeGPReleasePipelineModel(payload, name)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.admin.SaveGPReleasePipelineBody(r.Context(), name, version, raw); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "gp release not found"})
			return
		}
		if errors.Is(err, store.ErrComponentVersionNotDraft) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, store.ErrInvalidComposition) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		s.logger.Error("save gp release pipeline", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func normalizeGPReleasePipelineModel(raw []byte, gpName string) ([]byte, error) {
	var model map[string]any
	if err := json.Unmarshal(raw, &model); err != nil {
		return nil, err
	}
	if _, ok := model["schemaVersion"]; !ok {
		model["schemaVersion"] = gpcontent.SchemaVersionInline
	}
	model["name"] = gpName
	model["kind"] = "golden-path"
	return json.Marshal(model)
}
