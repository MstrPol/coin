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

var knownArtifactExts = map[string]struct{}{
	"json": {}, "sh": {}, "yaml": {}, "yml": {}, "md": {}, "txt": {},
}

// ArtifactMavenCoords maps GP content artifact key to Maven classifier + extension.
// Keys without a known extension (e.g. dockerfiles/Containerfile) keep the full
// dotted path as classifier and use a synthetic ext so Nexus maven-hosted accepts PUT.
func ArtifactMavenCoords(key string) (classifier, ext string) {
	normalized := strings.NewReplacer("/", ".", " ", "_").Replace(key)
	if i := strings.LastIndex(key, "."); i >= 0 {
		candidate := strings.ToLower(key[i+1:])
		if _, ok := knownArtifactExts[candidate]; ok {
			return normalized[:strings.LastIndex(normalized, ".")], candidate
		}
	}
	base := key
	if j := strings.LastIndex(key, "/"); j >= 0 {
		base = key[j+1:]
	}
	if strings.EqualFold(base, "Containerfile") {
		return normalized, "containerfile"
	}
	return normalized, ""
}

// ClassifierFromArtifactKey maps artifact key to Maven classifier (extension stripped).
func ClassifierFromArtifactKey(key string) string {
	classifier, _ := ArtifactMavenCoords(key)
	return classifier
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
