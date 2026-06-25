package store

import (
	"context"
	"encoding/json"
	"time"
)

type PlatformSettings struct {
	NexusMavenBase     string    `json:"nexusMavenBase"`
	NexusCredentialsID string    `json:"nexusCredentialsId"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func (s *Store) GetPlatformSettings(ctx context.Context) (PlatformSettings, error) {
	var row PlatformSettings
	err := s.pool.QueryRow(ctx, `
		SELECT nexus_maven_base, nexus_credentials_id, updated_at
		FROM platform_settings WHERE id = 1
	`).Scan(&row.NexusMavenBase, &row.NexusCredentialsID, &row.UpdatedAt)
	if err != nil {
		return PlatformSettings{}, err
	}
	return row, nil
}

func (s *Store) UpdatePlatformSettings(ctx context.Context, in PlatformSettings, actor string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE platform_settings
		SET nexus_maven_base = $1, nexus_credentials_id = $2, updated_at = now()
		WHERE id = 1
	`, in.NexusMavenBase, in.NexusCredentialsID)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(in)
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('update_platform_settings', 'platform_settings', 'global', $1, $2)
	`, nullIfEmpty(actor), payload)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
