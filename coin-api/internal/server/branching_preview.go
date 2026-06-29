package server

import (
	"encoding/json"
	"net/http"

	"coin.local/coin-executor/pkg/branching"
)

type branchingPreviewRequest struct {
	Model struct {
		Name     string                  `json:"name"`
		Branches []branching.BranchRuleDoc `json:"branches"`
	} `json:"model"`
	Scenarios []branchingPreviewScenario `json:"scenarios"`
}

type branchingPreviewScenario struct {
	ID             string   `json:"id"`
	Branch         string   `json:"branch"`
	TagName        string   `json:"tagName,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	RequestPublish bool     `json:"requestPublish,omitempty"`
}

func (s *Server) branchingModelPreview(w http.ResponseWriter, r *http.Request) {
	var req branchingPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	model, err := branching.ModelFromDoc(branching.ModelDoc{
		Name:     req.Model.Name,
		Branches: req.Model.Branches,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": branching.FormatPreviewError(err)})
		return
	}
	scenarios := make([]branching.PreviewScenario, 0, len(req.Scenarios))
	for _, sc := range req.Scenarios {
		scenarios = append(scenarios, branching.PreviewScenario{
			ID:             sc.ID,
			Branch:         sc.Branch,
			TagName:        sc.TagName,
			Tags:           sc.Tags,
			RequestPublish: sc.RequestPublish,
		})
	}
	result, err := branching.PreviewScenarios(model, scenarios)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": branching.FormatPreviewError(err)})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
