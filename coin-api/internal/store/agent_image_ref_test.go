package store

import (
	"testing"
)

func TestParseAgentImageRef(t *testing.T) {
	cases := []struct {
		image    string
		repo     string
		tag      string
		wantErr  bool
	}{
		{"nexus:8082/coin-docker/agent-30-06:1.2.0", "agent-30-06", "1.2.0", false},
		{"nexus:8082/coin-docker/coin-agent:0.1.0-draft", "coin-agent", "0.1.0-draft", false},
		{"coin-agent:2.0.0", "coin-agent", "2.0.0", false},
		{"nexus:8082/coin-docker/coin-agent:2.0.0@sha256:abc", "coin-agent", "2.0.0", false},
		{"nexus:8082/coin-docker/coin-agent", "", "", true},
		{"nexus:8082/coin-docker/coin-agent:latest", "", "", true},
		{"", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.image, func(t *testing.T) {
			parsed, err := ParseAgentImageRef(tc.image)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if parsed.Repository != tc.repo || parsed.Tag != tc.tag {
				t.Fatalf("got %#v want repo=%q tag=%q", parsed, tc.repo, tc.tag)
			}
		})
	}
}

func TestResolveAgentDraftVersion(t *testing.T) {
	meta := []byte(`{"image":"nexus:8082/coin-docker/coin-agent:1.2.0","digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`)
	ver, _, err := resolveAgentDraftVersion("coin-agent", "", meta)
	if err != nil {
		t.Fatal(err)
	}
	if ver != "1.2.0" {
		t.Fatalf("got %q", ver)
	}
	ver, _, err = resolveAgentDraftVersion("coin-agent", "1.2.0", meta)
	if err != nil || ver != "1.2.0" {
		t.Fatalf("CI path: %v ver=%q", err, ver)
	}
	_, _, err = resolveAgentDraftVersion("coin-agent", "9.9.9", meta)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
	_, _, err = resolveAgentDraftVersion("other-profile", "", meta)
	if err == nil {
		t.Fatal("expected profile mismatch")
	}
}
