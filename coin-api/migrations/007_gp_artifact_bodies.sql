-- +goose Up
CREATE TABLE gp_artifact_bodies (
    id              BIGSERIAL PRIMARY KEY,
    gp_release_id   BIGINT NOT NULL REFERENCES gp_releases(id) ON DELETE CASCADE,
    artifact_key    TEXT NOT NULL,
    body            BYTEA NOT NULL,
    sha256          TEXT NOT NULL,
    UNIQUE (gp_release_id, artifact_key)
);

CREATE INDEX idx_gp_artifact_bodies_release ON gp_artifact_bodies (gp_release_id);

-- +goose Down
DROP TABLE IF EXISTS gp_artifact_bodies;
