package gitea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTag(t *testing.T) {
	var gotTag string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/repos/coin/coin-jenkins-agents/tags" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotTag = body["tag_name"]
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := &Client{
		baseURL:      srv.URL,
		user:         "coin",
		password:     "coin",
		platformRepo: "coin/coin-jenkins-agents",
		httpClient:   srv.Client(),
	}
	if err := c.CreateTag(context.Background(), "go-app/v1.0.2"); err != nil {
		t.Fatal(err)
	}
	if gotTag != "go-app/v1.0.2" {
		t.Fatalf("tag %q", gotTag)
	}
}

func TestCreateTagConflictIsExists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()

	c := &Client{
		baseURL:      srv.URL,
		user:         "coin",
		password:     "coin",
		platformRepo: "coin/coin-jenkins-agents",
		httpClient:   srv.Client(),
	}
	if err := c.CreateTag(context.Background(), "go-app/v1.0.0"); err != ErrTagExists {
		t.Fatalf("expected ErrTagExists, got %v", err)
	}
}
