-- +goose Up
ALTER TABLE gp_profiles ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE gp_profiles DROP COLUMN IF EXISTS slots;

-- +goose Down
ALTER TABLE gp_profiles ADD COLUMN IF NOT EXISTS slots JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE gp_profiles DROP COLUMN IF EXISTS description;
