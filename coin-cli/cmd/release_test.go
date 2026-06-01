package cmd

import (
	"testing"
)

// ── bump ────────────────────────────────────────────────────────────────────

func TestBump_Patch(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1.0.0", "1.0.1"},
		{"1.4.3", "1.4.4"},
		{"0.0.0", "0.0.1"},
		{"10.20.30", "10.20.31"},
	}
	for _, c := range cases {
		got, err := bump(c.in, "patch")
		if err != nil {
			t.Errorf("bump(%q, patch): %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("bump(%q, patch) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBump_Minor(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1.0.0", "1.1.0"},
		{"1.4.3", "1.5.0"},   // patch сбрасывается
		{"0.0.9", "0.1.0"},
	}
	for _, c := range cases {
		got, err := bump(c.in, "minor")
		if err != nil {
			t.Errorf("bump(%q, minor): %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("bump(%q, minor) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBump_Major(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1.0.0", "2.0.0"},
		{"1.4.3", "2.0.0"},   // minor и patch сбрасываются
		{"0.9.9", "1.0.0"},
	}
	for _, c := range cases {
		got, err := bump(c.in, "major")
		if err != nil {
			t.Errorf("bump(%q, major): %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("bump(%q, major) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBump_Invalid(t *testing.T) {
	cases := []string{"1.0", "v1.0.0", "1.0.0-rc.1", "latest", ""}
	for _, c := range cases {
		_, err := bump(c, "patch")
		if err == nil {
			t.Errorf("bump(%q, patch): expected error, got nil", c)
		}
	}
}

// ── rcBaseVersion ───────────────────────────────────────────────────────────

func TestRcBaseVersion(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"1.5.0-rc.1", "1.5.0"},
		{"1.5.0-rc.2", "1.5.0"},
		{"2.0.0-rc.10", "2.0.0"},
		{"1.5.0", ""},           // финальная версия — не RC
		{"1.5.0-rc", ""},        // неполный формат
		{"v1.5.0-rc.1", ""},     // с префиксом — не матчит (prefix уже обрезан)
		{"", ""},
	}
	for _, c := range cases {
		got := rcBaseVersion(c.in)
		if got != c.want {
			t.Errorf("rcBaseVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── orNone ──────────────────────────────────────────────────────────────────

func TestOrNone(t *testing.T) {
	if got := orNone(""); got != "(нет тегов)" {
		t.Errorf("orNone(\"\") = %q, want (нет тегов)", got)
	}
	if got := orNone("v1.0.0"); got != "v1.0.0" {
		t.Errorf("orNone(\"v1.0.0\") = %q, want v1.0.0", got)
	}
}
