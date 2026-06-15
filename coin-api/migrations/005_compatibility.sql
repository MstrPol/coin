-- +goose Up
CREATE TABLE component_compatibility (
    id                      BIGSERIAL PRIMARY KEY,
    source_type             TEXT NOT NULL,
    source_name             TEXT NOT NULL,
    source_version_prefix   TEXT NOT NULL,
    requirements            JSONB NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_type, source_name, source_version_prefix)
);

-- +goose Down
DROP TABLE IF EXISTS component_compatibility;
