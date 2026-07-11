package report

import "testing"

func TestNormalizeResult(t *testing.T) {
	cases := map[string]string{
		"SUCCESS":  "success",
		"Failure":  "failure",
		"ABORTED":  "aborted",
		"unstable": "aborted",
	}
	for in, want := range cases {
		if got := NormalizeResult(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}
