package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coin.local/coin-api/internal/config"
)

func TestPrincipalHasRole(t *testing.T) {
	tests := []struct {
		roles []Role
		min   Role
		want  bool
	}{
		{[]Role{RoleAdmin}, RoleReader, true},
		{[]Role{RoleAdmin}, RolePublisher, true},
		{[]Role{RolePublisher}, RolePublisher, true},
		{[]Role{RolePublisher}, RoleAdmin, false},
		{[]Role{RoleReader}, RolePublisher, false},
		{[]Role{RoleReader}, RoleReader, true},
	}
	for _, tc := range tests {
		p := Principal{Roles: tc.roles}
		if got := p.Has(tc.min); got != tc.want {
			t.Fatalf("roles=%v min=%s got=%v want=%v", tc.roles, tc.min, got, tc.want)
		}
	}
}

func TestRequireRoleMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-API-Key", "reader-secret")
	rec := httptest.NewRecorder()
	AdminAuth(config.Config{
		AuthDisabled: false,
		ReaderAPIKey: "reader-secret",
	}, nil)(RequireRole(RolePublisher)(next)).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("reader publish: expected 403, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-API-Key", "pub-secret")
	rec = httptest.NewRecorder()
	AdminAuth(config.Config{
		AuthDisabled:    false,
		PublisherAPIKey: "pub-secret",
	}, nil)(RequireRole(RolePublisher)(next)).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("publisher publish: expected 200, got %d", rec.Code)
	}
}
