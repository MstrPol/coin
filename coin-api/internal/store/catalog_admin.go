package store

import (
	"context"
	"encoding/json"
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
	return CatalogPolicyRow{
		GPName:       policy.GPName,
		Latest:       policy.Latest,
		LatestCanary: policy.LatestCanary,
		Minimum:      policy.Minimum,
		Deprecated:   policy.Deprecated,
	}, nil
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
