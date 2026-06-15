package pin

import "testing"

func TestStripSnapshotVersion(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"1.0.5-snapshot.1": "1.0.5",
		"1.0.5":            "1.0.5",
		"2.0.0-snapshot.12": "2.0.0",
	}
	for in, want := range cases {
		if got := StripSnapshotVersion(in); got != want {
			t.Fatalf("StripSnapshotVersion(%q) = %q, want %q", in, got, want)
		}
	}
}
