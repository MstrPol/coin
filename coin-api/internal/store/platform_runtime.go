package store

import (
	"fmt"

	"coin.local/coin-api/internal/compatibility"
)

var gpSlotKeys = map[string]bool{
	"agent":           true,
	"branching-model": true,
}

var platformSlotKeys = map[string]bool{
	"lib": true,
}

func gpDraftCompositionSlots(agentStackName, branchingModelName string) []compatibility.CompositionSlot {
	return []compatibility.CompositionSlot{
		{Key: "agent", Type: "agent", Name: agentStackName},
		{Key: "branching-model", Type: "branching-model", Name: branchingModelName},
	}
}

func validateNewGPComposition(agentStackName, branchingModelName string, composition map[string]string) ([]compatibility.CompositionSlot, error) {
	if agentStackName == "" {
		return nil, fmt.Errorf("agentStackName is required")
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
		if key == "executor" || key == "gp-content" {
			return nil, fmt.Errorf("composition must not include %q slot", key)
		}
		if !gpSlotKeys[key] {
			return nil, fmt.Errorf("unknown composition key %q", key)
		}
	}
	for _, required := range []string{"agent", "branching-model"} {
		if composition[required] == "" {
			return nil, fmt.Errorf("composition.%s is required", required)
		}
	}
	return gpDraftCompositionSlots(agentStackName, branchingModelName), nil
}
