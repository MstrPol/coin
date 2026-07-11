package store

import "testing"

func TestMergeSlotsForValidation(t *testing.T) {
	gpSlots, err := validateNewGPComposition("coin-agent", "trunk-based", map[string]string{
		"agent":           "1.0.0",
		"branching-model": "1.0.0",
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(gpSlots) != 2 {
		t.Fatalf("expected 2 validation slots, got %d", len(gpSlots))
	}
}

func TestGpDraftCompositionSlots(t *testing.T) {
	slots := gpDraftCompositionSlots("coin-agent", "trunk-based")
	if slots[1].Name != "trunk-based" {
		t.Fatalf("expected branching-model name trunk-based, got %q", slots[1].Name)
	}
}

func TestValidateNewGPCompositionRejectsLib(t *testing.T) {
	_, err := validateNewGPComposition("coin-agent", "trunk-based", map[string]string{
		"agent":           "1.0.0",
		"branching-model": "1.0.0",
		"lib":             "1.0.0",
	})
	if err == nil {
		t.Fatal("expected lib in composition to be rejected")
	}
}
