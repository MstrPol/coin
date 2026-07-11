-- +goose Up
ALTER TABLE platform_settings ADD COLUMN IF NOT EXISTS runtime JSONB NOT NULL DEFAULT '{
  "agent": {"type": "agent", "name": "coin-agent", "version": "1.0.0"},
  "executor": {"type": "executor", "name": "coin-executor", "version": "1.0.0"},
  "lib": {"type": "lib", "name": "coin-lib", "version": "1.0.0"}
}'::jsonb;

-- +goose Down
ALTER TABLE platform_settings DROP COLUMN IF EXISTS runtime;
