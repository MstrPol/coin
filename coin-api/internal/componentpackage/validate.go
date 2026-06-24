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

	// PGOnlyRegistryComponentType is the first component type with draft/canary in PG only (Nexus on promote).
	PGOnlyRegistryComponentType = "branching-model"
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

// UsesPGOnlyCanaryRegistry reports whether register-package skips Nexus until promote.
func UsesPGOnlyCanaryRegistry(typ string) bool {
	return typ == PGOnlyRegistryComponentType
}

// IsContentRefV2Envelope reports whether raw JSON uses the v2 envelope (package optional).
func IsContentRefV2Envelope(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var probe struct {
		SchemaVersion int `json:"schemaVersion"`
	}
	if json.Unmarshal(raw, &probe) != nil {
		return false
	}
	return probe.SchemaVersion == ContentRefSchemaVersion
}

// IsContentRefV2 reports whether raw JSON is a full published v2 ref with package.url.
func IsContentRefV2(raw json.RawMessage) bool {
	return IsContentRefV2Envelope(raw) && HasPackageURL(raw)
}

// HasPackageURL reports whether content_ref v2 includes a non-empty package.url.
func HasPackageURL(raw json.RawMessage) bool {
	ref, err := parseContentRefV2Envelope(raw)
	if err != nil {
		return false
	}
	return strings.TrimSpace(ref.Package.URL) != ""
}

// IsRegisteredForCanary reports PG register completeness (manifest subset or full package ref).
func IsRegisteredForCanary(raw json.RawMessage) bool {
	if !IsContentRefV2Envelope(raw) {
		return false
	}
	if HasPackageURL(raw) {
		return true
	}
	ref, err := parseContentRefV2Envelope(raw)
	if err != nil {
		return false
	}
	return len(ref.Manifest) > 0
}

func parseContentRefV2Envelope(raw json.RawMessage) (ContentRefV2, error) {
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
	return ref, nil
}

func validateContentRefV2Envelope(raw json.RawMessage) (ContentRefV2, error) {
	ref, err := parseContentRefV2Envelope(raw)
	if err != nil {
		return ContentRefV2{}, err
	}
	if strings.TrimSpace(ref.Package.URL) != "" {
		if _, err := url.Parse(ref.Package.URL); err != nil {
			return ContentRefV2{}, fmt.Errorf("content_ref.package.url: %w", err)
		}
		if !sha256Pattern.MatchString(ref.Package.SHA256) {
			return ContentRefV2{}, errors.New("content_ref.package.sha256 must match sha256:<hex>")
		}
	} else if ref.Package.SHA256 != "" {
		return ContentRefV2{}, errors.New("content_ref.package.sha256 without package.url")
	}
	return ref, nil
}

// ManifestSubsetFromContentRef returns manifest from a v2 envelope when present.
func ManifestSubsetFromContentRef(raw json.RawMessage) map[string]any {
	ref, err := parseContentRefV2Envelope(raw)
	if err != nil || len(ref.Manifest) == 0 {
		return nil
	}
	return ref.Manifest
}

func ValidateContentRefV2(raw json.RawMessage) (ContentRefV2, error) {
	ref, err := validateContentRefV2Envelope(raw)
	if err != nil {
		return ContentRefV2{}, err
	}
	if strings.TrimSpace(ref.Package.URL) == "" {
		return ContentRefV2{}, errors.New("content_ref.package.url is required")
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
	if !IsContentRefV2Envelope(raw) {
		return nil
	}
	if _, err := validateContentRefV2Envelope(raw); err != nil {
		return err
	}
	return nil
}

func BuildContentRefV2PGOnly(manifestSubset map[string]any) (json.RawMessage, error) {
	ref := ContentRefV2{
		SchemaVersion: ContentRefSchemaVersion,
		Manifest:      manifestSubset,
	}
	raw, err := json.Marshal(ref)
	if err != nil {
		return nil, err
	}
	if _, err := validateContentRefV2Envelope(raw); err != nil {
		return nil, err
	}
	if len(manifestSubset) == 0 {
		return nil, errors.New("manifest subset is required for PG-only content_ref")
	}
	return raw, nil
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
