package server

import (
	"encoding/json"
	"net/http"

	"coin.local/coin-api/internal/gpcontent"
)

type gpContentPreviewRequest struct {
	ContentYAML   string `json:"contentYaml"`
	ComponentName string `json:"componentName"`
	Project       string `json:"project"`
	RegistryHost  string `json:"registryHost"`
}

func (s *Server) gpContentPreview(w http.ResponseWriter, r *http.Request) {
	var req gpContentPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	if req.ContentYAML == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "contentYaml is required"})
		return
	}
	doc, err := gpcontent.ParseDoc([]byte(req.ContentYAML))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	result := gpcontent.Preview(doc, gpcontent.PreviewOptions{
		ComponentName: req.ComponentName,
		Project:       req.Project,
		RegistryHost:  req.RegistryHost,
	})
	writeJSON(w, http.StatusOK, result)
}
