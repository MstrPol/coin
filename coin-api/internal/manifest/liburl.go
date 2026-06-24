package manifest

import (
	"os"

	"coin.local/coin-api/internal/nexus"
)

// LibZipURL builds the Nexus maven2 URL for a Jenkins Shared Library ZIP.
func LibZipURL(name, version string) string {
	base := os.Getenv("NEXUS_URL")
	if base == "" {
		base = "http://nexus:8081"
	}
	repo := nexus.MavenRepoForVersion(version)
	return nexus.MavenArtifactURL(base, repo, "coin.lib", name, version, "", "zip")
}
