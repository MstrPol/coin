-- +goose Up
-- Remove Jenkins lib from control plane: registry rows, GP composition pins, platform runtime column.

DELETE FROM component_artifact_bodies
WHERE component_version_id IN (
  SELECT cv.id FROM component_versions cv
  JOIN components c ON c.id = cv.component_id
  WHERE c.type = 'lib'
);

DELETE FROM component_versions
WHERE component_id IN (SELECT id FROM components WHERE type = 'lib');

DELETE FROM components WHERE type = 'lib';

DELETE FROM gp_composition WHERE component_type = 'lib';

ALTER TABLE platform_settings DROP COLUMN IF EXISTS runtime;

-- +goose Down
ALTER TABLE platform_settings ADD COLUMN IF NOT EXISTS runtime JSONB NOT NULL DEFAULT '{"lib": {"type": "lib", "name": "coin-lib", "version": "1.0.0"}}'::jsonb;
