package compatibility

import "testing"

func TestValidateGoAppComposition(t *testing.T) {
	slots := []CompositionSlot{
		{Key: "executor", Type: "executor", Name: "coin-executor"},
		{Key: "agent", Type: "agent", Name: "go"},
		{Key: "pipeline", Type: "pipeline", Name: "go-build"},
		{Key: "validate", Type: "validate", Name: "config"},
		{Key: "dockerfile", Type: "dockerfile", Name: "go-runtime"},
	}
	rules := []Rule{{
		SourceType:    "pipeline",
		SourceName:    "go-build",
		VersionPrefix: "2.1.",
		Requirements: map[string]Requirement{
			"executor": {Type: "executor", Name: "coin-executor", Min: "0.1.0", MaxExclusive: "0.2.0"},
			"agent":      {Type: "agent", Name: "go", Min: "1.22.0"},
		},
	}}

	ok := map[string]string{
		"executor": "0.1.0", "agent": "1.22.5", "pipeline": "2.1.0",
		"validate": "1.0.0", "dockerfile": "1.0.0",
	}
	if err := Validate(slots, ok, rules); err != nil {
		t.Fatalf("expected ok: %v", err)
	}

	bad := map[string]string{
		"executor": "0.2.0", "agent": "1.22.5", "pipeline": "2.1.0",
		"validate": "1.0.0", "dockerfile": "1.0.0",
	}
	if err := Validate(slots, bad, rules); err == nil {
		t.Fatal("expected executor compatibility error")
	}
}
