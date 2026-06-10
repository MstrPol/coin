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

INSERT INTO component_compatibility (source_type, source_name, source_version_prefix, requirements)
VALUES (
    'pipeline', 'go-build', '2.1.',
    '{
        "executor": {"type": "executor", "name": "coin-executor", "min": "0.1.0", "maxExclusive": "0.2.0"},
        "agent": {"type": "agent", "name": "go", "min": "1.22.0"}
    }'::jsonb
);

-- +goose Down
DROP TABLE IF EXISTS component_compatibility;
