package store

import "testing"

func TestValidateCanonicalGPSlots(t *testing.T) {
	valid := []GPProfileSlot{
		{Key: "jnlp", Type: "agent", Name: "jnlp"},
		{Key: "agent", Type: "agent", Name: "go"},
		{Key: "executor", Type: "executor", Name: "coin-executor"},
		{Key: "lib", Type: "lib", Name: "coin-lib"},
		{Key: "gp-content", Type: "gp-content", Name: "go-app"},
	}
	if err := ValidateCanonicalGPSlots(valid); err != nil {
		t.Fatalf("valid slots: %v", err)
	}
	if err := ValidateCanonicalGPSlots(valid[:4]); err == nil {
		t.Fatal("expected error for 4 slots")
	}
}
