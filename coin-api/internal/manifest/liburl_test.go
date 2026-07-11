package manifest

import "testing"

func TestLibZipURL(t *testing.T) {
	t.Setenv("NEXUS_URL", "http://nexus:8081")
	got := LibZipURL("coin-lib", "1.0.0")
	want := "http://nexus:8081/repository/maven-releases/coin/lib/coin-lib/1.0.0/coin-lib-1.0.0.zip"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
