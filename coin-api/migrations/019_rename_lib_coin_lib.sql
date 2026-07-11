-- +goose Up
-- Rename Jenkins Shared Library component platform-starter → coin-lib.
UPDATE components
SET name = 'coin-lib'
WHERE type = 'lib' AND name = 'platform-starter';

UPDATE gp_composition
SET component_name = 'coin-lib'
WHERE component_type = 'lib' AND component_name = 'platform-starter';

UPDATE gp_profiles gp
SET slots = sub.new_slots
FROM (
    SELECT
        p.name,
        jsonb_agg(
            CASE
                WHEN elem->>'key' = 'lib' AND elem->>'name' = 'platform-starter'
                    THEN jsonb_set(elem, '{name}', '"coin-lib"')
                ELSE elem
            END
            ORDER BY ord
        ) AS new_slots
    FROM gp_profiles p
    CROSS JOIN LATERAL jsonb_array_elements(p.slots) WITH ORDINALITY AS t(elem, ord)
    GROUP BY p.name
) sub
WHERE gp.name = sub.name;

-- +goose Down
-- Irreversible: coin-lib rename is not rolled back.
