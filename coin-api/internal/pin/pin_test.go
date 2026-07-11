package pin

import (
	"testing"
)

func TestParseAndSatisfies(t *testing.T) {
	p, err := Parse("~1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if !p.Satisfies("1.0.4") {
		t.Fatal("1.0.4 should satisfy ~1.0.0")
	}
	if p.Satisfies("1.1.0") {
		t.Fatal("1.1.0 should not satisfy ~1.0.0")
	}

	exact, _ := Parse("=1.0.0")
	if exact.Satisfies("1.0.1") {
		t.Fatal("exact pin mismatch")
	}
	if !exact.Satisfies("1.0.0") {
		t.Fatal("exact pin should match")
	}

	caret, _ := Parse("^1.0.0")
	if !caret.Satisfies("1.2.3") {
		t.Fatal("1.2.3 should satisfy ^1.0.0")
	}
	if caret.Satisfies("2.0.0") {
		t.Fatal("2.0.0 should not satisfy ^1.0.0")
	}
}

func TestSelectBest(t *testing.T) {
	candidates := []string{"1.0.0", "1.0.3", "1.0.1", "1.1.0"}
	p, _ := Parse("~1.0.0")
	got, err := p.SelectBest(candidates, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.0.3" {
		t.Fatalf("got %s", got)
	}
}

func TestTildeCaretPointer(t *testing.T) {
	if got := TildePointer("1.0.4"); got != "~1.0.0" {
		t.Fatalf("tilde pointer %q", got)
	}
	if got := CaretPointer("1.2.3"); got != "^1.0.0" {
		t.Fatalf("caret pointer %q", got)
	}
}
