package store

import (
	"context"
	"fmt"

	"coin.local/coin-api/internal/compatibility"
)

// ComponentPin is a type/name/version reference to a platform component.
type ComponentPin struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

var gpSlotKeys = map[string]bool{
	"agent":           true,
	"gp-content":      true,
	"branching-model": true,
}

var platformSlotKeys = map[string]bool{
	"lib": true,
}

func DerivedExecutorPin(agentName, agentVersion string) (ComponentPin, bool) {
	pin, err := executorPinForAgentStack(agentName, agentVersion)
	if err != nil {
		return ComponentPin{}, false
	}
	return pin, true
}

func executorPinForAgentStack(agentName, agentVersion string) (ComponentPin, error) {
	if agentName == "" || agentVersion == "" {
		return ComponentPin{}, fmt.Errorf("agent stack name and version are required")
	}
	return ComponentPin{Type: "executor", Name: "coin-executor", Version: agentVersion}, nil
}

func gpDraftCompositionSlots(agentStackName, gpContentName, branchingModelName string) []compatibility.CompositionSlot {
	return []compatibility.CompositionSlot{
		{Key: "agent", Type: "agent", Name: agentStackName},
		{Key: "gp-content", Type: "gp-content", Name: gpContentName},
		{Key: "branching-model", Type: "branching-model", Name: branchingModelName},
	}
}

func validateNewGPComposition(agentStackName, gpContentName, branchingModelName string, composition map[string]string) ([]compatibility.CompositionSlot, error) {
	if agentStackName == "" {
		return nil, fmt.Errorf("agentStackName is required")
	}
	if gpContentName == "" {
		return nil, fmt.Errorf("gpContentName is required")
	}
	if branchingModelName == "" {
		return nil, fmt.Errorf("branchingModelName is required")
	}
	if len(composition) == 0 {
		return nil, fmt.Errorf("composition is required")
	}
	for key := range composition {
		if platformSlotKeys[key] {
			return nil, fmt.Errorf("composition must not include platform slot %q", key)
		}
		if key == "executor" {
			return nil, fmt.Errorf("composition must not include standalone executor (bundled in agent stack)")
		}
		if !gpSlotKeys[key] {
			return nil, fmt.Errorf("unknown composition key %q", key)
		}
	}
	for _, required := range []string{"agent", "gp-content", "branching-model"} {
		if composition[required] == "" {
			return nil, fmt.Errorf("composition.%s is required", required)
		}
	}
	return gpDraftCompositionSlots(agentStackName, gpContentName, branchingModelName), nil
}

func isLegacyFullComposition(composition map[string]string) bool {
	legacyKeys := map[string]bool{"agent": true, "executor": true, "gp-content": true, "branching-model": true}
	for key := range legacyKeys {
		if _, ok := composition[key]; !ok {
			return false
		}
	}
	return len(composition) >= 4
}

func mergeSlotsForValidation(gpSlots, extraSlots []compatibility.CompositionSlot) []compatibility.CompositionSlot {
	out := make([]compatibility.CompositionSlot, 0, len(gpSlots)+len(extraSlots))
	out = append(out, gpSlots...)
	out = append(out, extraSlots...)
	return out
}

func mergeCompositionMaps(gpComp map[string]string, extra map[string]string) map[string]string {
	out := make(map[string]string, len(gpComp)+len(extra))
	for k, v := range extra {
		out[k] = v
	}
	for k, v := range gpComp {
		out[k] = v
	}
	return out
}

func (s *Store) agentVersionFromComposition(ctx context.Context, gpName, gpVersion string) (string, string, error) {
	var name, version string
	err := s.pool.QueryRow(ctx, `
		SELECT gc.component_name, gc.component_version
		FROM gp_composition gc
		JOIN gp_releases gr ON gr.id = gc.gp_release_id
		WHERE gr.name = $1 AND gr.version = $2 AND gc.component_type = 'agent'
	`, gpName, gpVersion).Scan(&name, &version)
	if err != nil {
		return "", "", err
	}
	return name, version, nil
}
