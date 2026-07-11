-- +goose Up
ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS group_id TEXT,
    ADD COLUMN IF NOT EXISTS artifact_id TEXT,
    ADD COLUMN IF NOT EXISTS git_repo_name TEXT,
    ADD COLUMN IF NOT EXISTS git_repo_url TEXT,
    ADD COLUMN IF NOT EXISTS gp_pin TEXT,
    ADD COLUMN IF NOT EXISTS version_pin TEXT,
    ADD COLUMN IF NOT EXISTS last_build_at TIMESTAMPTZ;

DROP TABLE IF EXISTS project_bindings;
DROP TABLE IF EXISTS scanner_state;

-- +goose Down
CREATE TABLE IF NOT EXISTS project_bindings (
    project_id    BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    gp_name       TEXT NOT NULL,
    gp_version    TEXT NOT NULL,
    git_url       TEXT,
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, gp_name, gp_version)
);

CREATE TABLE IF NOT EXISTS scanner_state (
    id         BIGSERIAL PRIMARY KEY,
    last_run   TIMESTAMPTZ,
    last_error TEXT
);

ALTER TABLE projects
    DROP COLUMN IF EXISTS last_build_at,
    DROP COLUMN IF EXISTS version_pin,
    DROP COLUMN IF EXISTS gp_pin,
    DROP COLUMN IF EXISTS git_repo_url,
    DROP COLUMN IF EXISTS git_repo_name,
    DROP COLUMN IF EXISTS artifact_id,
    DROP COLUMN IF EXISTS group_id;
