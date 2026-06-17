package store

import "testing"

func TestValidateCanonicalGPSlots(t *testing.T) {
	valid := []GPProfileSlot{
		{Key: "agent", Type: "agent", Name: "coin-agent"},
		{Key: "executor", Type: "executor", Name: "coin-executor"},
		{Key: "lib", Type: "lib", Name: "coin-lib"},
		{Key: "gp-content", Type: "gp-content", Name: "go-app"},
	}
	if err := ValidateCanonicalGPSlots(valid); err != nil {
		t.Fatalf("valid slots: %v", err)
	}
	if err := ValidateCanonicalGPSlots(valid[:3]); err == nil {
		t.Fatal("expected error for 3 slots")
	}
}
