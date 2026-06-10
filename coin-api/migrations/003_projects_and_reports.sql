-- +goose Up
CREATE TABLE projects (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE project_bindings (
    project_id    BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    gp_name       TEXT NOT NULL,
    gp_version    TEXT NOT NULL,
    git_url       TEXT,
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, gp_name, gp_version)
);

CREATE TABLE build_reports (
    id             BIGSERIAL PRIMARY KEY,
    project_id     BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    gp_name        TEXT NOT NULL,
    gp_version     TEXT NOT NULL,
    branch         TEXT,
    build_url      TEXT,
    result         TEXT NOT NULL,
    manifest_hash  TEXT,
    reported_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_build_reports_project ON build_reports(project_id, reported_at DESC);
CREATE INDEX idx_project_bindings_gp ON project_bindings(gp_name, gp_version);

-- +goose Down
DROP TABLE IF EXISTS build_reports;
DROP TABLE IF EXISTS project_bindings;
DROP TABLE IF EXISTS projects;
