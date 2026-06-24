package componentpackage

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

type ArtifactInput struct {
	Path   string
	SHA256 string
}

// InferFileRole maps artifact path to package.manifest role.
func InferFileRole(filePath string) string {
	base := path.Base(filePath)
	lower := strings.ToLower(filePath)
	switch {
	case strings.HasPrefix(lower, "schema/"):
		return "schema"
	case strings.HasPrefix(lower, "dockerfiles/") || base == "Dockerfile" || base == "Containerfile":
		return "containerfile"
	case strings.HasSuffix(lower, ".md"):
		return "docs"
	case strings.HasSuffix(lower, ".zip"):
		return "archive"
	case base == "content.yaml" || base == "model.yaml":
		return "primary"
	default:
		return "other"
	}
}

// BuildPackageManifestJSON builds Nexus package.manifest.json bytes from draft artifact bodies.
func BuildPackageManifestJSON(componentType, componentName, version string, artifacts []ArtifactInput) ([]byte, error) {
	if componentType == "" || componentName == "" || version == "" {
		return nil, fmt.Errorf("componentType, componentName and version are required")
	}
	if len(artifacts) == 0 {
		return nil, fmt.Errorf("package must contain at least one file")
	}
	files := make([]PackageFile, 0, len(artifacts))
	seen := make(map[string]struct{}, len(artifacts))
	for _, a := range artifacts {
		if a.Path == "" {
			return nil, fmt.Errorf("artifact path is required")
		}
		if !sha256Pattern.MatchString(a.SHA256) {
			return nil, fmt.Errorf("artifact %q: invalid sha256", a.Path)
		}
		if _, ok := seen[a.Path]; ok {
			return nil, fmt.Errorf("duplicate artifact path %q", a.Path)
		}
		seen[a.Path] = struct{}{}
		files = append(files, PackageFile{
			Path:   a.Path,
			SHA256: a.SHA256,
			Role:   InferFileRole(a.Path),
		})
	}
	manifest := PackageManifest{
		SchemaVersion: PackageManifestVersion,
		ComponentType: componentType,
		ComponentName: componentName,
		Version:       version,
		Files:         files,
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	if _, err := ValidatePackageManifest(raw); err != nil {
		return nil, err
	}
	return raw, nil
}
