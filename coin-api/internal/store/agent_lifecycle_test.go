package store

import (
	"encoding/json"
	"testing"
)

func TestValidateContentRefOnWriteForAgentSkipsValidation(t *testing.T) {
	bad := json.RawMessage(`{"schemaVersion":2,"packageUrl":"not-a-url"}`)
	if err := validateContentRefOnWriteForType("agent", bad); err != nil {
		t.Fatalf("agent must skip content_ref validation: %v", err)
	}
}

func TestValidateContentRefOnWriteForGPContentStillValidates(t *testing.T) {
	bad := json.RawMessage(`{
		"schemaVersion": 2,
		"package": {"sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	}`)
	if err := validateContentRefOnWriteForType("gp-content", bad); err == nil {
		t.Fatal("expected validation error for gp-content content_ref")
	}
}

func TestDerivedExecutorPin(t *testing.T) {
	pin, ok := DerivedExecutorPin("coin-agent", "1.2.0")
	if !ok {
		t.Fatal("expected derived pin")
	}
	if pin.Type != "executor" || pin.Name != "coin-executor" || pin.Version != "1.2.0" {
		t.Fatalf("unexpected pin: %#v", pin)
	}
	_, ok = DerivedExecutorPin("unknown-agent", "1.0.0")
	if ok {
		t.Fatal("expected no pin for unknown agent stack")
	}
}
