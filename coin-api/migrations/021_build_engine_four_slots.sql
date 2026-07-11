-- +goose Up
-- Five-slot profiles (jnlp + stack agent) → four-slot build-engine model (coin-agent only).
UPDATE gp_profiles gp
SET slots = jsonb_build_array(
    jsonb_build_object('key', 'agent', 'type', 'agent', 'name', 'coin-agent'),
    jsonb_build_object('key', 'executor', 'type', 'executor', 'name', 'coin-executor'),
    jsonb_build_object('key', 'lib', 'type', 'lib', 'name', 'coin-lib'),
    jsonb_build_object('key', 'gp-content', 'type', 'gp-content', 'name', gp.name)
)
FROM (
    SELECT p.name, jsonb_array_length(p.slots) AS slot_count
    FROM gp_profiles p
) src
WHERE gp.name = src.name
  AND src.slot_count <> 4;

-- +goose Down
-- Irreversible: legacy slot layout is not restored.
