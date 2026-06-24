package componentpackage

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	ContentRefSchemaVersion = 2
	PackageManifestVersion  = 1
)

var sha256Pattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

type PackageRef struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type ContentRefV2 struct {
	SchemaVersion int            `json:"schemaVersion"`
	Package       PackageRef     `json:"package"`
	Manifest      map[string]any `json:"manifest,omitempty"`
}

type PackageFile struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	Role      string `json:"role,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
}

type PackageManifest struct {
	SchemaVersion int           `json:"schemaVersion"`
	ComponentType string        `json:"componentType"`
	ComponentName string        `json:"componentName"`
	Version       string        `json:"version"`
	Files         []PackageFile `json:"files"`
}

// IsContentRefV2 reports whether raw JSON uses the v2 package envelope.
func IsContentRefV2(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var probe struct {
		SchemaVersion int             `json:"schemaVersion"`
		Package       json.RawMessage `json:"package"`
	}
	if json.Unmarshal(raw, &probe) != nil {
		return false
	}
	return probe.SchemaVersion == ContentRefSchemaVersion && len(probe.Package) > 0
}

func ValidateContentRefV2(raw json.RawMessage) (ContentRefV2, error) {
	var ref ContentRefV2
	if len(raw) == 0 {
		return ContentRefV2{}, errors.New("content_ref is required")
	}
	if err := json.Unmarshal(raw, &ref); err != nil {
		return ContentRefV2{}, fmt.Errorf("content_ref json: %w", err)
	}
	if ref.SchemaVersion != ContentRefSchemaVersion {
		return ContentRefV2{}, fmt.Errorf("content_ref schemaVersion must be %d", ContentRefSchemaVersion)
	}
	if strings.TrimSpace(ref.Package.URL) == "" {
		return ContentRefV2{}, errors.New("content_ref.package.url is required")
	}
	if _, err := url.Parse(ref.Package.URL); err != nil {
		return ContentRefV2{}, fmt.Errorf("content_ref.package.url: %w", err)
	}
	if !sha256Pattern.MatchString(ref.Package.SHA256) {
		return ContentRefV2{}, errors.New("content_ref.package.sha256 must match sha256:<hex>")
	}
	return ref, nil
}

func ValidatePackageManifest(raw json.RawMessage) (PackageManifest, error) {
	var manifest PackageManifest
	if len(raw) == 0 {
		return PackageManifest{}, errors.New("package manifest is required")
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return PackageManifest{}, fmt.Errorf("package manifest json: %w", err)
	}
	if manifest.SchemaVersion != PackageManifestVersion {
		return PackageManifest{}, fmt.Errorf("package manifest schemaVersion must be %d", PackageManifestVersion)
	}
	if manifest.ComponentType == "" || manifest.ComponentName == "" || manifest.Version == "" {
		return PackageManifest{}, errors.New("componentType, componentName and version are required")
	}
	if len(manifest.Files) == 0 {
		return PackageManifest{}, errors.New("package manifest files must not be empty")
	}
	seen := make(map[string]struct{}, len(manifest.Files))
	for _, f := range manifest.Files {
		if f.Path == "" {
			return PackageManifest{}, errors.New("package file path is required")
		}
		if !sha256Pattern.MatchString(f.SHA256) {
			return PackageManifest{}, fmt.Errorf("package file %q: invalid sha256", f.Path)
		}
		if _, ok := seen[f.Path]; ok {
			return PackageManifest{}, fmt.Errorf("duplicate package file path %q", f.Path)
		}
		seen[f.Path] = struct{}{}
	}
	return manifest, nil
}

// ValidateContentRefOnWrite validates content_ref when present on create/update.
// Legacy refs (artifactKey, gp-content manifest subset) pass through unchanged.
func ValidateContentRefOnWrite(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	if !IsContentRefV2(raw) {
		return nil
	}
	if _, err := ValidateContentRefV2(raw); err != nil {
		return err
	}
	return nil
}

func BuildContentRefV2(packageURL, packageSHA256 string, manifestSubset map[string]any) (json.RawMessage, error) {
	ref := ContentRefV2{
		SchemaVersion: ContentRefSchemaVersion,
		Package: PackageRef{
			URL:    packageURL,
			SHA256: packageSHA256,
		},
		Manifest: manifestSubset,
	}
	raw, err := json.Marshal(ref)
	if err != nil {
		return nil, err
	}
	if _, err := ValidateContentRefV2(raw); err != nil {
		return nil, err
	}
	return raw, nil
}
