package store

import "context"

func (s *Store) ManifestHash(ctx context.Context, gpName, version string) (string, error) {
	var hash *string
	err := s.pool.QueryRow(ctx, `
		SELECT manifest_hash FROM gp_releases WHERE name=$1 AND version=$2
	`, gpName, version).Scan(&hash)
	if err != nil {
		return "", err
	}
	if hash == nil {
		return "", nil
	}
	return *hash, nil
}
