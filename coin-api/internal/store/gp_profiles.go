package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/compatibility"
)

type GPProfileSlot struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type GPProfile struct {
	Name             string          `json:"name"`
	AgentStack       string          `json:"agentStack,omitempty"`
	DefaultLib       string          `json:"defaultLib,omitempty"`
	DefaultGPContent string          `json:"defaultGpContent,omitempty"`
	Slots            []GPProfileSlot `json:"slots"`
}

func (s *Store) profileSlots(ctx context.Context, name string) ([]compatibility.CompositionSlot, error) {
	var raw []byte
	err := s.pool.QueryRow(ctx, `SELECT slots FROM gp_profiles WHERE name=$1`, name).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedGP, name)
	}
	if err != nil {
		return nil, err
	}
	var slots []GPProfileSlot
	if err := json.Unmarshal(raw, &slots); err != nil {
		return nil, fmt.Errorf("gp profile slots: %w", err)
	}
	if len(slots) == 0 {
		return nil, fmt.Errorf("%w: %s (empty slots)", ErrUnsupportedGP, name)
	}
	out := make([]compatibility.CompositionSlot, len(slots))
	for i, slot := range slots {
		out[i] = compatibility.CompositionSlot{Key: slot.Key, Type: slot.Type, Name: slot.Name}
	}
	return out, nil
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

func agentStackFromCompositionSlots(slots []compatibility.CompositionSlot) string {
	for _, slot := range slots {
		if slot.Key == "agent" {
			return slot.Name
		}
	}
	return ""
}

func (s *Store) GetGPProfile(ctx context.Context, name string) (GPProfile, error) {
	slots, err := s.profileSlots(ctx, name)
	if err != nil {
		return GPProfile{}, err
	}
	out := GPProfile{
		Name:       name,
		AgentStack: agentStackFromCompositionSlots(slots),
		Slots:      make([]GPProfileSlot, len(slots)),
	}
	if out.AgentStack != "" {
		out.DefaultLib = "lib/coin-lib"
		out.DefaultGPContent = "gp-content/" + name
	}
	for i, slot := range slots {
		out.Slots[i] = GPProfileSlot{Key: slot.Key, Type: slot.Type, Name: slot.Name}
	}
	return out, nil
}

func (s *Store) CreateGPProfile(ctx context.Context, name string, slots []GPProfileSlot) error {
	if name == "" || len(slots) == 0 {
		return fmt.Errorf("name and slots are required")
	}
	raw, err := json.Marshal(slots)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO gp_profiles (name, slots) VALUES ($1, $2::jsonb)
	`, name, raw)
	if isUniqueViolation(err) {
		return ErrDuplicateGPProfile
	}
	return err
}

// CreateGPProfileByAgentStack creates a canonical four-slot profile (agentStack is ignored).
func (s *Store) CreateGPProfileByAgentStack(ctx context.Context, name, _ string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	return s.CreateGPProfile(ctx, name, CanonicalGPSlots(name))
}

var canonicalSlotKeys = []string{"agent", "executor", "lib", "gp-content"}

// ValidateCanonicalGPSlots ensures profile uses the four-component GP model.
func ValidateCanonicalGPSlots(slots []GPProfileSlot) error {
	if len(slots) != len(canonicalSlotKeys) {
		return fmt.Errorf("gp profile must have exactly %d slots", len(canonicalSlotKeys))
	}
	seen := make(map[string]bool, len(slots))
	for _, slot := range slots {
		if slot.Key == "" || slot.Type == "" || slot.Name == "" {
			return fmt.Errorf("slot key, type and name are required")
		}
		seen[slot.Key] = true
	}
	for _, key := range canonicalSlotKeys {
		if !seen[key] {
			return fmt.Errorf("missing slot %q", key)
		}
	}
	for _, slot := range slots {
		switch slot.Key {
		case "agent":
			if slot.Type != "agent" || slot.Name != "coin-agent" {
				return fmt.Errorf("agent slot must be agent/coin-agent")
			}
		case "executor":
			if slot.Type != "executor" || slot.Name != "coin-executor" {
				return fmt.Errorf("executor slot must be executor/coin-executor")
			}
		case "lib":
			if slot.Type != "lib" || slot.Name == "" {
				return fmt.Errorf("lib slot must be lib/{name}")
			}
		case "gp-content":
			if slot.Type != "gp-content" || slot.Name == "" {
				return fmt.Errorf("gp-content slot must be gp-content/{golden-path}")
			}
		default:
			return fmt.Errorf("unknown slot key %q", slot.Key)
		}
	}
	return nil
}

func (s *Store) CreateGPProfileWithDefaults(ctx context.Context, name string, slots []GPProfileSlot, actor string) error {
	if err := s.CreateGPProfile(ctx, name, slots); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO canary_policy (gp_name) VALUES ($1) ON CONFLICT DO NOTHING`, name); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('create_gp_profile', 'gp_profile', $1, $2, '{}'::jsonb)
	`, name, nullIfEmpty(actor))
	return err
}

func (s *Store) CreateGPProfileByAgentStackWithDefaults(ctx context.Context, name, agentStack, actor string) error {
	if err := s.CreateGPProfileByAgentStack(ctx, name, agentStack); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO canary_policy (gp_name) VALUES ($1) ON CONFLICT DO NOTHING`, name); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_log (action, entity_type, entity_key, actor, payload)
		VALUES ('create_gp_profile', 'gp_profile', $1, $2, '{}'::jsonb)
	`, name, nullIfEmpty(actor))
	return err
}

func (s *Store) CountGPProfiles(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM gp_profiles`).Scan(&n)
	return n, err
}

// SortProfileSlots returns a copy sorted by key (for stable API output).
func SortProfileSlots(slots []GPProfileSlot) []GPProfileSlot {
	out := append([]GPProfileSlot(nil), slots...)
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
