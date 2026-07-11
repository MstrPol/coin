package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coin.local/coin-api/internal/config"
)

func TestAdminKeyDisabled(t *testing.T) {
	cfg := config.Config{AuthDisabled: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	AdminAuth(cfg, nil)(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminKeyRequired(t *testing.T) {
	cfg := config.Config{AuthDisabled: false, AdminAPIKey: "admin-secret"}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	 w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	AdminAuth(cfg, nil)(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAdminKeyValidHeader(t *testing.T) {
	cfg := config.Config{AuthDisabled: false, AdminAPIKey: "admin-secret"}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-API-Key", "admin-secret")
	rec := httptest.NewRecorder()
	AdminAuth(cfg, nil)(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
