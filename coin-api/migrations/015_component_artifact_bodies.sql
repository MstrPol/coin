-- +goose Up
-- Artifact bodies for component versions (pipeline-bundle scripts, schemas, etc.)
CREATE TABLE IF NOT EXISTS component_artifact_bodies (
    component_version_id BIGINT NOT NULL REFERENCES component_versions(id) ON DELETE CASCADE,
    artifact_key         TEXT NOT NULL,
    body                 BYTEA NOT NULL,
    sha256               TEXT NOT NULL,
    PRIMARY KEY (component_version_id, artifact_key)
);

CREATE INDEX IF NOT EXISTS idx_component_artifact_bodies_version
    ON component_artifact_bodies (component_version_id);

-- +goose Down
DROP TABLE IF EXISTS component_artifact_bodies;
