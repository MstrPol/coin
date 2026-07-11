-- +goose Up
ALTER TABLE gp_releases
  ADD COLUMN image_registry_prefix TEXT NOT NULL DEFAULT 'localhost:8082/coin-docker',
  ADD COLUMN build_cache_enabled BOOLEAN NOT NULL DEFAULT true,
  ADD COLUMN artifact_repository_base TEXT NOT NULL DEFAULT 'http://nexus:8081/repository/maven-releases';

-- +goose Down
ALTER TABLE gp_releases
  DROP COLUMN IF EXISTS artifact_repository_base,
  DROP COLUMN IF EXISTS build_cache_enabled,
  DROP COLUMN IF EXISTS image_registry_prefix;
