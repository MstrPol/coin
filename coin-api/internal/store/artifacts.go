package store

import "context"

type ArtifactBody struct {
	Key    string
	Body   []byte
	SHA256 string
}

func (s *Store) ListArtifactBodies(ctx context.Context, gpName, gpVersion string) ([]ArtifactBody, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ab.artifact_key, ab.body, ab.sha256
		FROM gp_artifact_bodies ab
		JOIN gp_releases gr ON gr.id = ab.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2
	`, gpName, gpVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ArtifactBody
	for rows.Next() {
		var item ArtifactBody
		if err := rows.Scan(&item.Key, &item.Body, &item.SHA256); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
