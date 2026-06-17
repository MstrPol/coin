package build

import (
	"strings"
	"testing"
)

func TestBuildkitArgsOmitsEmptyTarget(t *testing.T) {
	args := buildkitArgs("/ws", "/ws/.coin", "Containerfile", Options{
		Workspace: "/ws",
		Target:    "",
	})
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "target=") {
		t.Fatalf("empty target must be omitted: %q", joined)
	}
}

func TestBuildkitArgsIncludesTarget(t *testing.T) {
	args := buildkitArgs("/ws", "/ws/.coin", "Containerfile", Options{
		Target: "runtime",
	})
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--opt target=runtime") {
		t.Fatalf("expected target opt: %q", joined)
	}
}
