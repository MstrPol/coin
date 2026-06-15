package manifest

import (
	"os"
	"strings"

	"coin.local/coin-api/internal/nexus"
)

// ContentArtifactURL builds the Nexus maven2 URL for a GP content artifact.
func ContentArtifactURL(gpName, gpVersion, artifactKey string) string {
	base := os.Getenv("NEXUS_URL")
	if base == "" {
		base = "http://nexus:8081"
	}
	ext := ""
	if i := strings.LastIndex(artifactKey, "."); i >= 0 {
		ext = artifactKey[i+1:]
	}
	classifier := nexus.ClassifierFromArtifactKey(artifactKey)
	repo := nexus.MavenRepoForVersion(gpVersion)
	return nexus.MavenArtifactURL(base, repo, "coin.gp.content", gpName, gpVersion, classifier, ext)
}
