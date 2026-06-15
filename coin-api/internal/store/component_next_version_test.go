package store

import (
	"strconv"
	"testing"
)

func TestAgentRevSuffixParsing(t *testing.T) {
	cases := []struct {
		v        string
		runtime  string
		rev      int
		ok       bool
	}{
		{"1.22-r1", "1.22", 1, true},
		{"1.22-r42", "1.22", 42, true},
		{"1.25-r3", "1.25", 3, true},
		{"1.22.5", "1.22", 0, false},
		{"1.22-r1", "1.25", 0, false},
	}
	for _, tc := range cases {
		m := agentRevSuffix.FindStringSubmatch(tc.v)
		if tc.ok {
			if m == nil || m[1] != tc.runtime {
				t.Fatalf("%s: expected match runtime=%s", tc.v, tc.runtime)
			}
			n, _ := strconv.Atoi(m[2])
			if n != tc.rev {
				t.Fatalf("%s: rev=%d want %d", tc.v, n, tc.rev)
			}
		} else if m != nil && m[1] == tc.runtime {
			t.Fatalf("%s: unexpected match", tc.v)
		}
	}
}
