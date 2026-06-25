-- +goose Up
-- Backfill agent pin for GP releases created under two-slot model (no agent row).
INSERT INTO gp_composition (gp_release_id, component_type, component_name, component_version)
SELECT gr.id,
  'agent',
  COALESCE(ps.runtime->'agent'->>'name', 'coin-agent'),
  COALESCE(ps.runtime->'agent'->>'version', '1.0.0')
FROM gp_releases gr
CROSS JOIN platform_settings ps
WHERE ps.id = 1
  AND NOT EXISTS (
    SELECT 1 FROM gp_composition gc
    WHERE gc.gp_release_id = gr.id AND gc.component_type = 'agent'
  )
ON CONFLICT DO NOTHING;

UPDATE platform_settings
SET runtime = jsonb_build_object(
  'lib', COALESCE(runtime->'lib', '{"type": "lib", "name": "coin-lib", "version": "1.0.0"}'::jsonb)
)
WHERE id = 1;

-- +goose Down
UPDATE platform_settings
SET runtime = runtime || jsonb_build_object(
  'agent', '{"type": "agent", "name": "coin-agent", "version": "1.0.0"}'::jsonb,
  'executor', '{"type": "executor", "name": "coin-executor", "version": "1.0.0"}'::jsonb
)
WHERE id = 1;
