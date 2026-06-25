-- +goose Up
-- Local pilot publishes coin-executor@1.0.0; migration 025 default 0.1.0 breaks draft validation.
UPDATE platform_settings
SET runtime = jsonb_set(runtime, '{executor,version}', '"1.0.0"')
WHERE runtime->'executor'->>'version' = '0.1.0';

-- +goose Down
UPDATE platform_settings
SET runtime = jsonb_set(runtime, '{executor,version}', '"0.1.0"')
WHERE runtime->'executor'->>'version' = '1.0.0';
