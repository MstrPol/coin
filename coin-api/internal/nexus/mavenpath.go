package nexus

import (
	"fmt"
	"strings"
)

const (
	MavenReleases  = "maven-releases"
	MavenSnapshots = "maven-snapshots"
)

// MavenRepoForVersion picks hosted repo by semver (-SNAPSHOT → snapshots).
func MavenRepoForVersion(version string) string {
	if strings.Contains(version, "SNAPSHOT") {
		return MavenSnapshots
	}
	return MavenReleases
}

// MavenArtifactURL builds a maven2 layout download URL.
func MavenArtifactURL(baseURL, repo, groupID, artifactID, version, classifier, ext string) string {
	base := strings.TrimRight(baseURL, "/")
	groupPath := strings.ReplaceAll(groupID, ".", "/")
	file := fmt.Sprintf("%s-%s", artifactID, version)
	if classifier != "" {
		file += "-" + classifier
	}
	if ext != "" {
		file += "." + ext
	}
	return fmt.Sprintf("%s/repository/%s/%s/%s/%s/%s", base, repo, groupPath, artifactID, version, file)
}

// MavenRepoPath returns repo-relative path (for PUT).
func MavenRepoPath(groupID, artifactID, version, classifier, ext string) string {
	groupPath := strings.ReplaceAll(groupID, ".", "/")
	file := fmt.Sprintf("%s-%s", artifactID, version)
	if classifier != "" {
		file += "-" + classifier
	}
	if ext != "" {
		file += "." + ext
	}
	return fmt.Sprintf("%s/%s/%s/%s", groupPath, artifactID, version, file)
}

// ClassifierFromArtifactKey maps artifact key to Maven classifier (extension stripped).
func ClassifierFromArtifactKey(key string) string {
	repl := strings.NewReplacer("/", ".", " ", "_").Replace(key)
	if i := strings.LastIndex(repl, "."); i > 0 {
		return repl[:i]
	}
	return repl
}

// ImmutableConflict reports Nexus maven-hosted 400 when the asset already exists.
func ImmutableConflict(statusCode int, body string) bool {
	if statusCode != 400 {
		return false
	}
	lower := strings.ToLower(body)
	return strings.Contains(lower, "already exists") ||
		strings.Contains(lower, "redeploy is not allowed") ||
		strings.Contains(lower, "cannot be updated")
}
