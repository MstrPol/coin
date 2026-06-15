-- +goose Up
-- Legacy 6-slot profiles → canonical 4-component model (agent stack from agent slot).
UPDATE gp_profiles gp
SET slots = jsonb_build_array(
    jsonb_build_object('key', 'jnlp', 'type', 'agent', 'name', 'jnlp'),
    jsonb_build_object('key', 'agent', 'type', 'agent', 'name', src.agent_name),
    jsonb_build_object('key', 'executor', 'type', 'executor', 'name', 'coin-executor'),
    jsonb_build_object('key', 'pipeline-bundle', 'type', 'pipeline-bundle', 'name', src.agent_name)
)
FROM (
    SELECT
        p.name,
        COALESCE(
            (
                SELECT elem->>'name'
                FROM jsonb_array_elements(p.slots) AS elem
                WHERE elem->>'key' = 'agent'
                LIMIT 1
            ),
            'go'
        ) AS agent_name,
        jsonb_array_length(p.slots) AS slot_count
    FROM gp_profiles p
) src
WHERE gp.name = src.name
  AND src.slot_count <> 4;

-- +goose Down
-- Irreversible: legacy slot layout is not restored.
