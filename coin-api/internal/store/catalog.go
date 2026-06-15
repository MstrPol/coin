package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/catalog"
)

func (s *Store) GetCatalogPolicy(ctx context.Context, gpName string) (catalog.Policy, error) {
	var policy catalog.Policy
	var deprecated []byte
	err := s.pool.QueryRow(ctx, `
		SELECT gp_name, latest, COALESCE(latest_canary, ''), minimum, deprecated
		FROM catalog_policy WHERE gp_name = $1
	`, gpName).Scan(&policy.GPName, &policy.Latest, &policy.LatestCanary, &policy.Minimum, &deprecated)
	if errors.Is(err, pgx.ErrNoRows) {
		return catalog.Policy{GPName: gpName, Deprecated: []string{}}, nil
	}
	if err != nil {
		return catalog.Policy{}, err
	}
	_ = json.Unmarshal(deprecated, &policy.Deprecated)
	if policy.Deprecated == nil {
		policy.Deprecated = []string{}
	}
	return policy, nil
}
