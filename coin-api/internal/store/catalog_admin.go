package store

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/mod/semver"

	"coin.local/coin-api/internal/pin"
)

type CatalogPolicyRow struct {
	GPName       string   `json:"gpName"`
	Latest       string   `json:"latest"`
	LatestCanary string   `json:"latestCanary"`
	Minimum      string   `json:"minimum"`
	Deprecated   []string `json:"deprecated"`
}

func (s *Store) GetCatalogPolicyRow(ctx context.Context, gpName string) (CatalogPolicyRow, error) {
	policy, err := s.GetCatalogPolicy(ctx, gpName)
	if err != nil {
		return CatalogPolicyRow{}, err
	}
	dep := policy.Deprecated
	if dep == nil {
		dep = []string{}
	}
	return CatalogPolicyRow{
		GPName:       policy.GPName,
		Latest:       policy.Latest,
		LatestCanary: policy.LatestCanary,
		Minimum:      policy.Minimum,
		Deprecated:   dep,
	}, nil
}

func (s *Store) ValidateCatalogPolicyUpdate(ctx context.Context, gpName, latest, latestCanary, minimum string, deprecated []string) error {
	published, err := s.ListPublishedGPVersions(ctx, gpName)
	if err != nil {
		return err
	}
	pubSet := make(map[string]struct{}, len(published))
	for _, v := range published {
		pubSet[v] = struct{}{}
	}
	ensurePublished := func(label, version string, required bool) error {
		if version == "" {
			if required {
				return fmt.Errorf("%s is required", label)
			}
			return nil
		}
		if pin.IsSnapshotVersion(version) {
			return fmt.Errorf("%s cannot be a snapshot version", label)
		}
		if _, ok := pubSet[version]; !ok {
			return fmt.Errorf("%s version %s is not published", label, version)
		}
		return nil
	}
	if err := ensurePublished("latest", latest, true); err != nil {
		return err
	}
	if err := ensurePublished("minimum", minimum, true); err != nil {
		return err
	}
	if err := ensurePublished("latestCanary", latestCanary, false); err != nil {
		return err
	}
	for _, v := range deprecated {
		if err := ensurePublished("deprecated", v, true); err != nil {
			return err
		}
	}
	if minimum != "" && latest != "" && semver.Compare(normSemver(minimum), normSemver(latest)) > 0 {
		return fmt.Errorf("minimum cannot be greater than latest")
	}
	return nil
}

func (s *Store) UpdateCatalogPolicy(ctx context.Context, gpName, latest, latestCanary, minimum string, deprecated []string, actor string) error {
	if deprecated == nil {
		deprecated = []string{}
	}
	depJSON, _ := json.Marshal(deprecated)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO catalog_policy (gp_name, latest, latest_canary, minimum, deprecated)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		ON CONFLICT (gp_name) DO UPDATE SET
			latest = EXCLUDED.latest,
			latest_canary = EXCLUDED.latest_canary,
			minimum = EXCLUDED.minimum,
			deprecated = EXCLUDED.deprecated
	`, gpName, latest, nullIfEmpty(latestCanary), minimum, depJSON)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('update_catalog', 'catalog_policy', $1, $2, $3)
	`, gpName, nullIfEmpty(actor), depJSON)
	return err
}
