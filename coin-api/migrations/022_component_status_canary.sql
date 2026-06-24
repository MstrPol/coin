-- +goose Up
ALTER TYPE component_status ADD VALUE IF NOT EXISTS 'canary' AFTER 'draft';

-- +goose Down
-- PostgreSQL does not support removing enum values; irreversible.
