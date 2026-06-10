package nexus

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublishManifestBlobAndPointer(t *testing.T) {
	var blobPath, pointerPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(r.URL.Path, "/manifest/blobs/") {
			blobPath = r.URL.Path
			if !strings.HasSuffix(blobPath, ".json") {
				t.Errorf("blob path: %s", blobPath)
			}
		}
		if strings.Contains(r.URL.Path, "/pointers/go-app/") {
			pointerPath = r.URL.Path
			var ptr PointerDoc
			if err := json.Unmarshal(body, &ptr); err != nil {
				t.Fatal(err)
			}
			if ptr.ManifestHash == "" || ptr.BlobURL == "" || ptr.ResolvedVersion != "1.0.0" {
				t.Fatalf("pointer: %#v", ptr)
			}
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := &Client{
		baseURL:    srv.URL,
		repository: "coin-manifests",
		httpClient: srv.Client(),
	}
	doc := map[string]any{
		"manifestHash": "sha256:deadbeef",
		"goldenPath":   map[string]string{"name": "go-app", "version": "1.0.0"},
	}
	_, err := c.PublishManifest(context.Background(), "go-app", "1.0.0", "=1.0.0", doc, "sha256:deadbeef")
	if err != nil {
		t.Fatal(err)
	}
	if blobPath == "" || pointerPath == "" {
		t.Fatalf("blob=%q pointer=%q", blobPath, pointerPath)
	}
	if !strings.Contains(pointerPath, "pointers/go-app/") {
		t.Fatalf("unexpected pointer path %s", pointerPath)
	}
}

func TestEncodePinForPath(t *testing.T) {
	if got := encodePinForPath("=1.0.0"); got != "%3D1.0.0" {
		t.Fatalf("got %q", got)
	}
	if got := encodePinForPath("~1.0.0"); got != "%7E1.0.0" {
		t.Fatalf("got %q", got)
	}
	if got := encodePinForPath("^1.0.0"); got != "%5E1.0.0" {
		t.Fatalf("got %q", got)
	}
}
