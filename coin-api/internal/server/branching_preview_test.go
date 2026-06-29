package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBranchingModelPreview_smoke(t *testing.T) {
	t.Parallel()

	body := `{
		"model": {
			"name": "trunk-based",
			"branches": [
				{
					"name": "feature",
					"pattern": "^feature/(?P<jira>[A-Z][A-Z0-9]*-\\d+)(?:-.+)?$",
					"versioning": {"template": "v{base}-{jira}-snapshot-{n}"},
					"publish": false
				},
				{
					"name": "release",
					"pattern": "^release/(?P<jira>[A-Z][A-Z0-9]*-\\d+)(?:-.+)?$",
					"versioning": {"template": "v{base}-{jira}-rc-{n}"},
					"publish": true
				}
			]
		},
		"scenarios": [
			{"id": "f1", "branch": "feature/PROJ-101", "requestPublish": true},
			{"id": "r1", "branch": "release/PROJ-404", "requestPublish": true}
		]
	}`

	s := &Server{}
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/branching-models/preview", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.branchingModelPreview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Results []struct {
			ID             string `json:"id"`
			MatchedRule    string `json:"matchedRule"`
			PublishOutcome string `json:"publishOutcome"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Results) != 2 {
		t.Fatalf("results=%+v", out.Results)
	}
	if out.Results[0].PublishOutcome != "denied" || out.Results[1].PublishOutcome != "allowed" {
		t.Fatalf("unexpected outcomes: %+v", out.Results)
	}
}
