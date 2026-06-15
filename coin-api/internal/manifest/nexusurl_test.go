package manifest

import (
	"os"
	"testing"
)

func TestRuntimeNexusURLRewritesLocalhost(t *testing.T) {
	t.Setenv("NEXUS_URL", "http://nexus:8081")
	got := RuntimeNexusURL("http://localhost:8081/repository/maven-releases/coin/pipeline-bundle/go/1.0.1/go-1.0.1.zip")
	want := "http://nexus:8081/repository/maven-releases/coin/pipeline-bundle/go/1.0.1/go-1.0.1.zip"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRuntimeNexusURLPreservesInternalHost(t *testing.T) {
	os.Unsetenv("NEXUS_URL")
	got := RuntimeNexusURL("http://nexus:8081/repository/maven-releases/coin/executor/coin-executor/1.0.0/coin-executor-1.0.0.bin")
	if got != "http://nexus:8081/repository/maven-releases/coin/executor/coin-executor/1.0.0/coin-executor-1.0.0.bin" {
		t.Fatalf("unexpected rewrite: %q", got)
	}
}
