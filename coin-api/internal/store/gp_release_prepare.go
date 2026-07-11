package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"coin.local/coin-api/internal/compatibility"
)

type preparedGPRelease struct {
	storeSlots        []compatibility.CompositionSlot
	validateSlots     []compatibility.CompositionSlot
	mergedComposition map[string]string
}

func (s *Store) prepareGPRelease(ctx context.Context, in PublishGPReleaseInput) (preparedGPRelease, error) {
	if err := s.gpProfileExists(ctx, in.Name); err != nil {
		return preparedGPRelease{}, err
	}

	gpSlots, err := validateNewGPComposition(in.AgentStackName, in.BranchingModelName, in.Composition)
	if err != nil {
		return preparedGPRelease{}, fmt.Errorf("%w: %v", ErrInvalidComposition, err)
	}

	merged := make(map[string]string, len(in.Composition))
	for k, v := range in.Composition {
		merged[k] = v
	}

	return preparedGPRelease{
		storeSlots:        gpSlots,
		validateSlots:     gpSlots,
		mergedComposition: merged,
	}, nil
}

func insertGPComposition(ctx context.Context, tx pgx.Tx, releaseID int64, slots []compatibility.CompositionSlot, composition map[string]string) error {
	for _, slot := range slots {
		ver := composition[slot.Key]
		_, err := tx.Exec(ctx, `
			INSERT INTO gp_composition (gp_release_id, component_type, component_name, component_version)
			VALUES ($1, $2, $3, $4)
		`, releaseID, slot.Type, slot.Name, ver)
		if err != nil {
			return fmt.Errorf("composition insert: %w", err)
		}
	}
	return nil
}

func validateGPReleaseComposition(
	s *Store,
	ctx context.Context,
	prep preparedGPRelease,
	rules []compatibility.Rule,
	resolveMode func(string) ComponentResolveMode,
) error {
	if err := compatibility.Validate(prep.validateSlots, prep.mergedComposition, rules); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidComposition, err)
	}
	for _, slot := range prep.validateSlots {
		ver := prep.mergedComposition[slot.Key]
		mode := resolveMode(slot.Type)
		ok, err := s.componentVersionResolvable(ctx, slot.Type, slot.Name, ver, mode)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%w: %s/%s@%s", ErrComponentNotFound, slot.Type, slot.Name, ver)
		}
	}
	return nil
}
