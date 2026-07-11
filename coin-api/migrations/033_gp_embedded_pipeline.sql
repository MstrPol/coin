-- +goose Up
CREATE TABLE IF NOT EXISTS gp_release_pipeline_bodies (
    gp_release_id   BIGINT PRIMARY KEY REFERENCES gp_releases(id) ON DELETE CASCADE,
    schema_version  INT NOT NULL DEFAULT 3,
    body            JSONB NOT NULL,
    sha256          TEXT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

DELETE FROM gp_composition WHERE component_type = 'gp-content';

DELETE FROM component_artifact_bodies cab
USING component_versions cv, components c
WHERE cab.component_version_id = cv.id
  AND cv.component_id = c.id
  AND c.type = 'gp-content';

DELETE FROM component_versions cv
USING components c
WHERE cv.component_id = c.id
  AND c.type = 'gp-content';

DELETE FROM components WHERE type = 'gp-content';

-- +goose Down
INSERT INTO components (type, name, description)
SELECT 'gp-content', name, description FROM gp_profiles
ON CONFLICT DO NOTHING;

DROP TABLE IF EXISTS gp_release_pipeline_bodies;
