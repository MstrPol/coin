-- +goose Up
-- Four-slot GP profiles → five-slot with branching-model (GBM-0.2).
UPDATE gp_profiles gp
SET slots = gp.slots || jsonb_build_array(
    jsonb_build_object(
        'key', 'branching-model',
        'type', 'branching-model',
        'name', CASE gp.name
            WHEN 'go-lib' THEN 'semver-tag'
            WHEN 'java-maven-app' THEN 'semver-tag'
            ELSE 'trunk-based'
        END
    )
)
WHERE jsonb_array_length(gp.slots) = 4
  AND NOT EXISTS (
    SELECT 1 FROM jsonb_array_elements(gp.slots) e WHERE e->>'key' = 'branching-model'
  );

-- +goose Down
-- Irreversible: branching-model slot is not removed on down.
