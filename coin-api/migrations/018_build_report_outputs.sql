-- +goose Up
ALTER TABLE build_reports ADD COLUMN IF NOT EXISTS outputs JSONB NOT NULL DEFAULT '[]'::jsonb;

-- +goose Down
ALTER TABLE build_reports DROP COLUMN IF EXISTS outputs;
