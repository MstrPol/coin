package store

import "testing"

func TestComponentVersionEntityKey(t *testing.T) {
	got := componentVersionEntityKey("agent", "coin-agent", "0.1.0-draft")
	want := "agent/coin-agent@0.1.0-draft"
	if got != want {
		t.Fatalf("entity key: got %q want %q", got, want)
	}
}
