package docs

import (
	_ "embed"
	"net/http"
	"os"
)

//go:embed swagger.html
var swaggerHTML []byte

// Handler serves OpenAPI spec and Swagger UI (no auth).
type Handler struct {
	openAPI []byte
}

// NewHandler loads openapi/v1.yaml from COIN_OPENAPI_PATH or common locations.
func NewHandler() *Handler {
	candidates := []string{
		os.Getenv("COIN_OPENAPI_PATH"),
		"openapi/v1.yaml",
		"/usr/share/coin/openapi/v1.yaml",
	}
	h := &Handler{}
	for _, path := range candidates {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err == nil {
			h.openAPI = data
			break
		}
	}
	return h
}

func (h *Handler) ServeOpenAPI(w http.ResponseWriter, _ *http.Request) {
	if len(h.openAPI) == 0 {
		http.Error(w, "openapi spec not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(h.openAPI)
}

func (h *Handler) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/docs" && r.URL.Path != "/docs/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(swaggerHTML)
}
