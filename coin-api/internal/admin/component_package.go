package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"coin.local/coin-api/internal/componentpackage"
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
	draftBodies := make([]componentpackage.DraftArtifact, 0, len(bodies))
	for _, b := range bodies {
		draftBodies = append(draftBodies, componentpackage.DraftArtifact{
			Path: b.Key, Body: b.Body, SHA256: b.SHA256,
		})
	}
	return componentpackage.ValidateDraftPackage(typ, name, version, draftBodies), nil
}

func (s *Service) RegisterComponentPackage(ctx context.Context, typ, name, version string, req RegisterComponentPackageRequest) (RegisterComponentPackageResult, error) {
	if s.nexus == nil {
		return RegisterComponentPackageResult{}, fmt.Errorf("nexus client not configured")
	}
	bodies, err := s.store.ListComponentArtifactBodies(ctx, typ, name, version)
	if err != nil {
		return RegisterComponentPackageResult{}, err
	}
	if len(bodies) == 0 {
		return RegisterComponentPackageResult{}, fmt.Errorf("draft has no artifact bodies to publish")
	}
	draftBodies := make([]componentpackage.DraftArtifact, 0, len(bodies))
	for _, b := range bodies {
		draftBodies = append(draftBodies, componentpackage.DraftArtifact{
			Path: b.Key, Body: b.Body, SHA256: b.SHA256,
		})
	}
	if v := componentpackage.ValidateDraftPackage(typ, name, version, draftBodies); !v.Valid {
		return RegisterComponentPackageResult{}, fmt.Errorf("package validation failed")
	}

	inputs := make([]componentpackage.ArtifactInput, 0, len(bodies))
	for _, b := range bodies {
		inputs = append(inputs, componentpackage.ArtifactInput{Path: b.Key, SHA256: b.SHA256})
	}
	manifestRaw, err := componentpackage.BuildPackageManifestJSON(typ, name, version, inputs)
	if err != nil {
		return RegisterComponentPackageResult{}, err
	}

	for _, b := range bodies {
		if _, err := s.nexus.UploadComponentPackageFile(ctx, typ, name, version, b.Key, b.Body); err != nil {
			return RegisterComponentPackageResult{}, fmt.Errorf("upload %q: %w", b.Key, err)
		}
	}
	packageURL, err := s.nexus.UploadComponentPackageManifest(ctx, typ, name, version, manifestRaw)
	if err != nil {
		return RegisterComponentPackageResult{}, fmt.Errorf("upload package manifest: %w", err)
	}
	sum := sha256.Sum256(manifestRaw)
	packageSHA := "sha256:" + hex.EncodeToString(sum[:])

	contentRef, err := componentpackage.BuildContentRefV2(packageURL, packageSHA, req.Manifest)
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
