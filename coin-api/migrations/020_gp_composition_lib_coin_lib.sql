-- +goose Up
-- Idempotent: fix lib slot name in published GP compositions (if 019 ran before gp_composition update).
UPDATE gp_composition
SET component_name = 'coin-lib'
WHERE component_type = 'lib' AND component_name = 'platform-starter';

-- +goose Down
-- Irreversible.
