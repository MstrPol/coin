package store

import (
	"testing"

	"coin.local/coin-api/internal/compatibility"
)

func TestIsLegacyFullComposition(t *testing.T) {
	legacy := map[string]string{
		"agent":           "1.0.0",
		"executor":        "0.1.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	}
	if !isLegacyFullComposition(legacy) {
		t.Fatal("expected legacy 4-slot composition")
	}

	threePin := map[string]string{
		"agent":           "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	}
	if isLegacyFullComposition(threePin) {
		t.Fatal("three-pin composition must not be legacy")
	}
}

func TestMergeCompositionMaps(t *testing.T) {
	gp := map[string]string{
		"agent":           "1.0.0",
		"gp-content":      "2.0.0",
		"branching-model": "1.0.0",
	}
	extra := map[string]string{
		"executor": "0.1.0",
	}
	merged := mergeCompositionMaps(gp, extra)
	if len(merged) != 4 {
		t.Fatalf("expected 4 keys, got %d", len(merged))
	}
	if merged["executor"] != "0.1.0" || merged["gp-content"] != "2.0.0" {
		t.Fatalf("unexpected merge: %#v", merged)
	}
}

func TestMergeSlotsForValidation(t *testing.T) {
	gpSlots, err := validateNewGPComposition("coin-agent", "go-app", "trunk-based", map[string]string{
		"agent":           "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	execPin, _ := executorPinForAgentStack("coin-agent", "1.0.0")
	executorSlot := compatibility.CompositionSlot{Key: "executor", Type: execPin.Type, Name: execPin.Name}
	all := mergeSlotsForValidation(gpSlots, []compatibility.CompositionSlot{executorSlot})
	if len(all) != 4 {
		t.Fatalf("expected 4 validation slots, got %d", len(all))
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
