-- +goose Up
CREATE TABLE gp_profiles (
    name       TEXT PRIMARY KEY,
    slots      JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS gp_profiles;
