package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type GPProfile struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (s *Store) gpProfileExists(ctx context.Context, name string) error {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT 1 FROM gp_profiles WHERE name=$1`, name).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %s", ErrUnsupportedGP, name)
	}
	return err
}

func (s *Store) ListGPNames(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT name FROM gp_profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func (s *Store) GetGPProfile(ctx context.Context, name string) (GPProfile, error) {
	var out GPProfile
	var desc *string
	err := s.pool.QueryRow(ctx, `
		SELECT name, description, created_at FROM gp_profiles WHERE name=$1
	`, name).Scan(&out.Name, &desc, &out.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return GPProfile{}, fmt.Errorf("%w: %s", ErrUnsupportedGP, name)
	}
	if err != nil {
		return GPProfile{}, err
	}
	if desc != nil {
		out.Description = *desc
	}
	return out, nil
}

func (s *Store) CreateGPProfile(ctx context.Context, name, description string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	var desc any
	if description != "" {
		desc = description
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO gp_profiles (name, description) VALUES ($1, $2)
	`, name, desc)
	if isUniqueViolation(err) {
		return ErrDuplicateGPProfile
	}
	return err
}

func (s *Store) CreateGPProfileWithDefaults(ctx context.Context, name, description, actor string) error {
	if err := s.CreateGPProfile(ctx, name, description); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO canary_policy (gp_name) VALUES ($1) ON CONFLICT DO NOTHING`, name); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]any{"description": description})
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('create_gp_profile', 'gp_profile', $1, $2, $3::jsonb)
	`, name, nullIfEmpty(actor), payload)
	return err
}

func (s *Store) CountGPProfiles(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM gp_profiles`).Scan(&n)
	return n, err
}
