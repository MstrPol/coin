package store

import "testing"

func TestBumpSemver(t *testing.T) {
	cases := []struct {
		current string
		bump    string
		want    string
	}{
		{"1.0.0", "patch", "1.0.1"},
		{"1.0.9", "patch", "1.0.10"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "major", "2.0.0"},
		{"v2.4.1", "patch", "2.4.2"},
	}
	for _, tc := range cases {
		got, err := bumpSemver(tc.current, tc.bump)
		if err != nil {
			t.Fatalf("%s %s: %v", tc.current, tc.bump, err)
		}
		if got != tc.want {
			t.Fatalf("%s %s: got %s want %s", tc.current, tc.bump, got, tc.want)
		}
	}
}

func TestLatestPublishedSemver(t *testing.T) {
	items := []ComponentVersionListItem{
		{Version: "1.0.0", Status: "published"},
		{Version: "1.1.0", Status: "published"},
		{Version: "2.0.0", Status: "draft"},
		{Version: "not-semver", Status: "published"},
	}
	if got := latestPublishedSemver(items); got != "1.1.0" {
		t.Fatalf("latest=%s want 1.1.0", got)
	}
	if got := latestPublishedSemver(nil); got != "" {
		t.Fatalf("empty list: got %q", got)
	}
}
