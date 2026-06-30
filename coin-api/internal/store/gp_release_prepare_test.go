package store

import "testing"

func TestMergeSlotsForValidation(t *testing.T) {
	gpSlots, err := validateNewGPComposition("coin-agent", "go-app", "trunk-based", map[string]string{
		"agent":           "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(gpSlots) != 3 {
		t.Fatalf("expected 3 validation slots, got %d", len(gpSlots))
	}
}

func TestGpDraftCompositionSlotsDecoupled(t *testing.T) {
	slots := gpDraftCompositionSlots("coin-agent", "go-app", "trunk-based")
	if slots[1].Name != "go-app" {
		t.Fatalf("expected gp-content name go-app, got %q", slots[1].Name)
	}
}

func TestValidateNewGPCompositionRejectsLib(t *testing.T) {
	_, err := validateNewGPComposition("coin-agent", "go-app", "trunk-based", map[string]string{
		"agent":           "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
		"lib":             "1.0.0",
	})
	if err == nil {
		t.Fatal("expected lib in composition to be rejected")
	}
}
