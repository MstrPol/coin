package store

import "testing"

func TestAllowedComponentStatuses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		mode    ComponentResolveMode
		want    []string
		check   string
		allowed bool
	}{
		{ComponentResolveStable, []string{"published"}, "published", true},
		{ComponentResolveStable, nil, "canary", false},
		{ComponentResolveStable, nil, "draft", false},
		{ComponentResolveDraft, []string{"published", "draft"}, "draft", true},
		{ComponentResolveDraft, nil, "published", true},
		{ComponentResolveDraft, nil, "canary", true},
	}
	for _, tc := range cases {
		got := allowedComponentStatuses(tc.mode)
		if tc.want != nil {
			if len(got) != len(tc.want) {
				t.Fatalf("mode %q: got %v want %v", tc.mode, got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("mode %q: got %v want %v", tc.mode, got, tc.want)
				}
			}
		}
		if componentStatusAllowed(tc.check, tc.mode) != tc.allowed {
			t.Fatalf("mode %q status %q allowed=%v", tc.mode, tc.check, tc.allowed)
		}
	}
}

func TestComponentResolveModeForGPDraftEdit(t *testing.T) {
	t.Parallel()
	if componentResolveModeForGPDraftEdit("agent") != ComponentResolveStable {
		t.Fatal("agent must be stable-only in GP draft")
	}
	if componentResolveModeForGPDraftEdit("branching-model") != ComponentResolveDraft {
		t.Fatal("branching-model allows draft in GP draft")
	}
}
