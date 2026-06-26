package store

import (
	"encoding/json"
	"testing"
)

func TestEncodePromoteBlockers(t *testing.T) {
	t.Parallel()

	blockers := []CompositionPinBlocker{
		{Type: "gp-content", Name: "go-app", Version: "1.0.1-draft", Status: "draft"},
		{Type: "branching-model", Name: "trunk-based", Version: "2.0.0-draft", Status: "draft"},
	}
	raw := encodePromoteBlockers(blockers)

	var payload struct {
		BlockingPins []CompositionPinBlocker `json:"blockingPins"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.BlockingPins) != 2 {
		t.Fatalf("blockingPins len = %d", len(payload.BlockingPins))
	}
	if payload.BlockingPins[0].Type != "gp-content" || payload.BlockingPins[0].Status != "draft" {
		t.Fatalf("unexpected first blocker: %+v", payload.BlockingPins[0])
	}
}

func TestErrGPCompositionHasDraftPins(t *testing.T) {
	t.Parallel()
	if ErrGPCompositionHasDraftPins.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}
