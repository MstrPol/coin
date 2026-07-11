package server

import (
	"testing"

	"coin.local/coin-api/internal/version"
)

func TestReadyPayloadIncludesVersion(t *testing.T) {
	t.Parallel()

	body := readyPayload("ready", "")
	if body["status"] != "ready" {
		t.Fatalf("status = %q", body["status"])
	}
	if body["version"] != version.Version {
		t.Fatalf("version = %q, want %q", body["version"], version.Version)
	}
	if _, ok := body["reason"]; ok {
		t.Fatal("reason should be omitted when empty")
	}

	notReady := readyPayload("not ready", "database")
	if notReady["reason"] != "database" {
		t.Fatalf("reason = %q", notReady["reason"])
	}
}
