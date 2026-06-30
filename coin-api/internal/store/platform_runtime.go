package store

import (
	"fmt"

	"coin.local/coin-api/internal/compatibility"
)

var gpSlotKeys = map[string]bool{
	"agent":           true,
	"gp-content":      true,
	"branching-model": true,
}

var platformSlotKeys = map[string]bool{
	"lib": true,
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
