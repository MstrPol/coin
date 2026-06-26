package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"coin.local/coin-api/internal/componentpackage"
	"coin.local/coin-api/internal/store"
)

type RegisterComponentPackageRequest struct {
	Manifest map[string]any `json:"manifest"`
	Actor    string         `json:"actor"`
}

type RegisterComponentPackageResult struct {
	Type          string          `json:"type"`
	Name          string          `json:"name"`
	Version       string          `json:"version"`
	PackageURL    string          `json:"packageUrl"`
	PackageSHA256 string          `json:"packageSha256"`
	FilesUploaded int             `json:"filesUploaded"`
	ContentRef    json.RawMessage `json:"contentRef"`
}

type ValidateComponentPackageResult = componentpackage.ValidateDraftResult

func (s *Service) ValidateComponentPackage(ctx context.Context, typ, name, version string) (ValidateComponentPackageResult, error) {
	bodies, err := s.store.ListComponentArtifactBodies(ctx, typ, name, version)
	if err != nil {
		return ValidateComponentPackageResult{}, err
	}
	return validateDraftBodies(typ, name, version, bodies), nil
}

func (s *Service) RegisterComponentPackage(ctx context.Context, typ, name, version string, req RegisterComponentPackageRequest) (RegisterComponentPackageResult, error) {
	bodies, err := s.store.ListComponentArtifactBodies(ctx, typ, name, version)
	if err != nil {
		return RegisterComponentPackageResult{}, err
	}
	if len(bodies) == 0 {
		return RegisterComponentPackageResult{}, fmt.Errorf("draft has no artifact bodies to publish")
	}
	if v := validateDraftBodies(typ, name, version, bodies); !v.Valid {
		return RegisterComponentPackageResult{}, fmt.Errorf("package validation failed")
	}

	if typ == componentpackage.PGOnlyRegistryComponentType {
		return s.registerComponentPackagePGOnly(ctx, typ, name, version, req.Manifest, bodies)
	}
	return s.registerComponentPackageNexus(ctx, typ, name, version, req.Manifest, bodies)
}

func validateDraftBodies(typ, name, version string, bodies []store.ComponentArtifactBody) componentpackage.ValidateDraftResult {
	draftBodies := make([]componentpackage.DraftArtifact, 0, len(bodies))
	for _, b := range bodies {
		draftBodies = append(draftBodies, componentpackage.DraftArtifact{
			Path: b.Key, Body: b.Body, SHA256: b.SHA256,
		})
	}
	return componentpackage.ValidateDraftPackage(typ, name, version, draftBodies)
}

func (s *Service) registerComponentPackagePGOnly(ctx context.Context, typ, name, version string, manifestSubset map[string]any, bodies []store.ComponentArtifactBody) (RegisterComponentPackageResult, error) {
	contentRef, err := componentpackage.BuildContentRefV2PGOnly(manifestSubset)
	if err != nil {
		return RegisterComponentPackageResult{}, fmt.Errorf("build content_ref: %w", err)
	}
	if err := s.store.UpdateComponentVersionContentRef(ctx, typ, name, version, contentRef); err != nil {
		return RegisterComponentPackageResult{}, err
	}
	return RegisterComponentPackageResult{
		Type:          typ,
		Name:          name,
		Version:       version,
		FilesUploaded: len(bodies),
		ContentRef:    contentRef,
	}, nil
}

func (s *Service) registerComponentPackageNexus(ctx context.Context, typ, name, version string, manifestSubset map[string]any, bodies []store.ComponentArtifactBody) (RegisterComponentPackageResult, error) {
	if s.nexus == nil {
		return RegisterComponentPackageResult{}, fmt.Errorf("nexus client not configured")
	}
	packageURL, packageSHA, err := s.uploadComponentPackageToNexus(ctx, typ, name, version, bodies)
	if err != nil {
		return RegisterComponentPackageResult{}, err
	}
	contentRef, err := componentpackage.BuildContentRefV2(packageURL, packageSHA, manifestSubset)
	if err != nil {
		return RegisterComponentPackageResult{}, fmt.Errorf("build content_ref: %w", err)
	}
	if err := s.store.UpdateComponentVersionContentRef(ctx, typ, name, version, contentRef); err != nil {
		return RegisterComponentPackageResult{}, err
	}
	return RegisterComponentPackageResult{
		Type:          typ,
		Name:          name,
		Version:       version,
		PackageURL:    packageURL,
		PackageSHA256: packageSHA,
		FilesUploaded: len(bodies),
		ContentRef:    contentRef,
	}, nil
}

func (s *Service) uploadComponentPackageToNexus(ctx context.Context, typ, name, version string, bodies []store.ComponentArtifactBody) (packageURL, packageSHA string, err error) {
	if s.nexus == nil {
		return "", "", fmt.Errorf("nexus client not configured")
	}
	inputs := make([]componentpackage.ArtifactInput, 0, len(bodies))
	for _, b := range bodies {
		inputs = append(inputs, componentpackage.ArtifactInput{Path: b.Key, SHA256: b.SHA256})
	}
	manifestRaw, err := componentpackage.BuildPackageManifestJSON(typ, name, version, inputs)
	if err != nil {
		return "", "", err
	}
	for _, b := range bodies {
		if _, err := s.nexus.UploadComponentPackageFile(ctx, typ, name, version, b.Key, b.Body); err != nil {
			return "", "", fmt.Errorf("upload %q: %w", b.Key, err)
		}
	}
	packageURL, err = s.nexus.UploadComponentPackageManifest(ctx, typ, name, version, manifestRaw)
	if err != nil {
		return "", "", fmt.Errorf("upload package manifest: %w", err)
	}
	sum := sha256.Sum256(manifestRaw)
	return packageURL, "sha256:" + hex.EncodeToString(sum[:]), nil
}

func (s *Service) publishComponentFromDraft(ctx context.Context, typ, name, version, actor string) (store.ComponentVersionRow, error) {
	detail, err := s.store.GetComponentVersionDetail(ctx, typ, name, version)
	if err != nil {
		return store.ComponentVersionRow{}, err
	}
	if detail.Status != "draft" && detail.Status != "canary" {
		return store.ComponentVersionRow{}, store.ErrComponentVersionNotDraft
	}
	if componentpackage.IsRegisteredForCanary(detail.ContentRef) || componentpackage.HasPackageURL(detail.ContentRef) {
		if componentpackage.HasPackageURL(detail.ContentRef) {
			return s.store.PromoteComponentToPublished(ctx, typ, name, version, actor)
		}
	}
	bodies, err := s.store.ListComponentArtifactBodiesForVersion(ctx, typ, name, version)
	if err != nil {
		return store.ComponentVersionRow{}, err
	}
	if len(bodies) == 0 {
		return store.ComponentVersionRow{}, fmt.Errorf("draft has no artifact bodies to publish")
	}
	packageURL, packageSHA, err := s.uploadComponentPackageToNexus(ctx, typ, name, version, bodies)
	if err != nil {
		return store.ComponentVersionRow{}, err
	}
	manifestSubset := componentpackage.ManifestSubsetFromContentRef(detail.ContentRef)
	contentRef, err := componentpackage.BuildContentRefV2(packageURL, packageSHA, manifestSubset)
	if err != nil {
		return store.ComponentVersionRow{}, fmt.Errorf("build content_ref: %w", err)
	}
	return s.store.PromoteComponentToPublishedWithContentRef(ctx, typ, name, version, contentRef, actor)
}
