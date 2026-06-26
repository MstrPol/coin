-- +goose Up
-- platform-native-lifecycle: component canary status superseded by GP-level canary only.

UPDATE component_versions
SET status = 'published'::component_status
WHERE status = 'canary'::component_status
  AND content_ref IS NOT NULL
  AND content_ref::text LIKE '%"url"%';

UPDATE component_versions
SET status = 'draft'::component_status
WHERE status = 'canary'::component_status;

-- +goose Down
-- Irreversible: cannot distinguish migrated rows from original draft/published.
