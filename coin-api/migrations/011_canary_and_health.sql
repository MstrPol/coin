-- +goose Up
ALTER TABLE catalog_policy ADD COLUMN IF NOT EXISTS latest_canary TEXT;

CREATE TYPE canary_mode AS ENUM ('default', 'canary', 'stable');

ALTER TABLE projects ADD COLUMN IF NOT EXISTS canary_mode canary_mode NOT NULL DEFAULT 'default';

CREATE TABLE canary_policy (
    gp_name                       TEXT PRIMARY KEY,
    enabled                       BOOLEAN NOT NULL DEFAULT false,
    canary_percent                INT NOT NULL DEFAULT 0 CHECK (canary_percent >= 0 AND canary_percent <= 100),
    degraded_threshold_pct        INT NOT NULL DEFAULT 10 CHECK (degraded_threshold_pct >= 0 AND degraded_threshold_pct <= 100),
    critical_threshold_pct        INT NOT NULL DEFAULT 25 CHECK (critical_threshold_pct >= 0 AND critical_threshold_pct <= 100),
    critical_consecutive_failures INT NOT NULL DEFAULT 3 CHECK (critical_consecutive_failures >= 1)
);

ALTER TABLE build_reports ADD COLUMN IF NOT EXISTS channel TEXT;
ALTER TABLE build_reports ADD COLUMN IF NOT EXISTS requested_pin TEXT;
ALTER TABLE build_reports ADD COLUMN IF NOT EXISTS failed_stage TEXT;
ALTER TABLE build_reports ADD COLUMN IF NOT EXISTS resolved_version TEXT;

CREATE INDEX IF NOT EXISTS idx_build_reports_gp_health
    ON build_reports (gp_name, gp_version, channel, reported_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_build_reports_gp_health;
ALTER TABLE build_reports DROP COLUMN IF EXISTS resolved_version;
ALTER TABLE build_reports DROP COLUMN IF EXISTS failed_stage;
ALTER TABLE build_reports DROP COLUMN IF EXISTS requested_pin;
ALTER TABLE build_reports DROP COLUMN IF EXISTS channel;
DROP TABLE IF EXISTS canary_policy;
ALTER TABLE projects DROP COLUMN IF EXISTS canary_mode;
DROP TYPE IF EXISTS canary_mode;
ALTER TABLE catalog_policy DROP COLUMN IF EXISTS latest_canary;
