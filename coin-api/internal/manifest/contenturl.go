package manifest

import (
	"os"

	"coin.local/coin-api/internal/nexus"
)

// ContentArtifactURL builds the Nexus maven2 URL for a GP content artifact.
func ContentArtifactURL(gpName, gpVersion, artifactKey string) string {
	base := os.Getenv("NEXUS_URL")
	if base == "" {
		base = "http://nexus:8081"
	}
	classifier, ext := nexus.ArtifactMavenCoords(artifactKey)
	repo := nexus.MavenRepoForVersion(gpVersion)
	return nexus.MavenArtifactURL(base, repo, "coin.gp.content", gpName, gpVersion, classifier, ext)
}
