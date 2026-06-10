package manifest

import (
	"fmt"
	"os"
)

// ContentArtifactURL builds the Nexus URL for a GP content artifact (PF-15 publishes bytes there).
func ContentArtifactURL(gpName, gpVersion, artifactKey string) string {
	base := os.Getenv("NEXUS_URL")
	if base == "" {
		base = "http://nexus:8081"
	}
	repo := os.Getenv("NEXUS_MANIFEST_REPO")
	if repo == "" {
		repo = "coin-manifests"
	}
	return fmt.Sprintf("%s/repository/%s/content/%s/%s/%s", base, repo, gpName, gpVersion, artifactKey)
}
