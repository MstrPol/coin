-- +goose Up
CREATE TABLE gp_profiles (
    name       TEXT PRIMARY KEY,
    slots      JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO gp_profiles (name, slots) VALUES (
    'go-app',
    '[
        {"key":"executor","type":"executor","name":"coin-executor"},
        {"key":"agent","type":"agent","name":"go"},
        {"key":"pipeline","type":"pipeline","name":"go-build"},
        {"key":"validate","type":"validate","name":"config"},
        {"key":"dockerfile","type":"dockerfile","name":"go-runtime"},
        {"key":"orchestration","type":"orchestration","name":"coin-pipeline"}
    ]'::jsonb
) ON CONFLICT (name) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS gp_profiles;
