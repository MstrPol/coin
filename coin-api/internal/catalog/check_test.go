package catalog

import "testing"

func TestCheckMinimum(t *testing.T) {
	p := Policy{GPName: "go-app", Minimum: "1.0.0", Deprecated: []string{}}
	if _, err := Check(p, "0.9.0"); err == nil {
		t.Fatal("expected below minimum error")
	}
	if _, err := Check(p, "1.0.0"); err != nil {
		t.Fatalf("1.0.0 should pass: %v", err)
	}
}

func TestCheckDeprecated(t *testing.T) {
	p := Policy{GPName: "go-app", Minimum: "1.0.0", Deprecated: []string{"1.0.0"}}
	w, err := Check(p, "1.0.0")
	if err != nil {
		t.Fatalf("deprecated should not error: %v", err)
	}
	if w == "" {
		t.Fatal("expected warning")
	}
}
