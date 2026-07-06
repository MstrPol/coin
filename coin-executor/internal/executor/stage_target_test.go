package executor

import "testing"

func TestInferStageTarget(t *testing.T) {
	tests := []struct {
		kind, output, buildType, want string
	}{
		{"run", "validate", "", "validate"},
		{"run", "test", "", "test"},
		{"run", "image", "", "runtime"},
		{"build", "", "image", "runtime"},
		{"build", "", "artifact", "artifact"},
		{"build", "", "liquibase-image", "runtime"},
	}
	for _, tc := range tests {
		if got := inferStageTarget(tc.kind, tc.output, tc.buildType); got != tc.want {
			t.Fatalf("inferStageTarget(%q, %q, %q) = %q, want %q", tc.kind, tc.output, tc.buildType, got, tc.want)
		}
	}
}
