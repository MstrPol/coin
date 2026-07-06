package store

import "testing"

func TestValidateNewGPCompositionRequiresAgentStackName(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"branching-model": "1.0.0",
	}
	if _, err := validateNewGPComposition("", "trunk-based", comp); err == nil {
		t.Fatal("expected error for empty agentStackName")
	}
}

func TestValidateNewGPCompositionTwoPin(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"branching-model": "1.0.0",
	}
	slots, err := validateNewGPComposition("coin-agent", "trunk-based", comp)
	if err != nil {
		t.Fatalf("valid composition: %v", err)
	}
	if len(slots) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(slots))
	}
	if slots[0].Key != "agent" || slots[0].Name != "coin-agent" {
		t.Fatalf("agent slot: %#v", slots[0])
	}
}

func TestValidateNewGPCompositionRejectsExecutorKey(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"executor":        "1.0.0",
		"branching-model": "1.0.0",
	}
	if _, err := validateNewGPComposition("coin-agent", "trunk-based", comp); err == nil {
		t.Fatal("expected error for standalone executor key")
	}
}

func TestValidateNewGPCompositionRejectsLibKey(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"lib":             "1.0.0",
		"branching-model": "1.0.0",
	}
	if _, err := validateNewGPComposition("coin-agent", "trunk-based", comp); err == nil {
		t.Fatal("expected error for lib key in GP composition")
	}
}

func TestValidateNewGPCompositionRejectsGPContentKey(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	}
	if _, err := validateNewGPComposition("coin-agent", "trunk-based", comp); err == nil {
		t.Fatal("expected error for gp-content key in GP composition")
	}
}
