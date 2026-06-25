package store

import "testing"

func TestValidateNewGPCompositionRequiresAgentStackName(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	}
	if _, err := validateNewGPComposition("", "go-app", "trunk-based", comp); err == nil {
		t.Fatal("expected error for empty agentStackName")
	}
}

func TestValidateNewGPCompositionThreePin(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	}
	slots, err := validateNewGPComposition("coin-agent", "go-app", "trunk-based", comp)
	if err != nil {
		t.Fatalf("valid composition: %v", err)
	}
	if len(slots) != 3 {
		t.Fatalf("expected 3 slots, got %d", len(slots))
	}
	if slots[0].Key != "agent" || slots[0].Name != "coin-agent" {
		t.Fatalf("agent slot: %#v", slots[0])
	}
}

func TestValidateNewGPCompositionRejectsExecutorKey(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"executor":        "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	}
	if _, err := validateNewGPComposition("coin-agent", "go-app", "trunk-based", comp); err == nil {
		t.Fatal("expected error for standalone executor key")
	}
}

func TestValidateNewGPCompositionRejectsLibKey(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"lib":             "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	}
	if _, err := validateNewGPComposition("coin-agent", "go-app", "trunk-based", comp); err == nil {
		t.Fatal("expected error for lib key in GP composition")
	}
}

func TestValidateNewGPCompositionDecoupledFromProfileName(t *testing.T) {
	comp := map[string]string{
		"agent":           "1.0.0",
		"gp-content":      "1.0.0",
		"branching-model": "1.0.0",
	}
	slots, err := validateNewGPComposition("coin-agent", "go-app", "trunk-based", comp)
	if err != nil {
		t.Fatalf("valid composition: %v", err)
	}
	if slots[1].Name != "go-app" {
		t.Fatalf("gp-content slot name = go-app, got %q", slots[1].Name)
	}
}

func TestExecutorPinForAgentStack(t *testing.T) {
	pin, err := executorPinForAgentStack("coin-agent", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if pin.Name != "coin-executor" || pin.Version != "1.0.0" {
		t.Fatalf("unexpected pin: %#v", pin)
	}
}
