-- +goose Up
CREATE TYPE component_status AS ENUM ('draft', 'published', 'deprecated', 'retired');

CREATE TABLE components (
    id          BIGSERIAL PRIMARY KEY,
    type        TEXT NOT NULL,
    name        TEXT NOT NULL,
    UNIQUE (type, name)
);

CREATE TABLE component_versions (
    id            BIGSERIAL PRIMARY KEY,
    component_id  BIGINT NOT NULL REFERENCES components(id),
    version       TEXT NOT NULL,
    status        component_status NOT NULL DEFAULT 'published',
    metadata      JSONB NOT NULL DEFAULT '{}',
    content_ref   JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (component_id, version)
);

CREATE TABLE gp_releases (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    version         TEXT NOT NULL,
    status          component_status NOT NULL DEFAULT 'published',
    manifest_hash   TEXT,
    manifest_url    TEXT,
    git_export_tag  TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (name, version)
);

CREATE TABLE gp_composition (
    gp_release_id       BIGINT NOT NULL REFERENCES gp_releases(id) ON DELETE CASCADE,
    component_type      TEXT NOT NULL,
    component_name      TEXT NOT NULL,
    component_version   TEXT NOT NULL,
    PRIMARY KEY (gp_release_id, component_type, component_name)
);

CREATE TABLE catalog_policy (
    gp_name   TEXT PRIMARY KEY,
    latest    TEXT,
    minimum   TEXT,
    deprecated JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX idx_gp_releases_name_version ON gp_releases(name, version);

-- +goose Down
DROP TABLE IF EXISTS catalog_policy;
DROP TABLE IF EXISTS gp_composition;
DROP TABLE IF EXISTS gp_releases;
DROP TABLE IF EXISTS component_versions;
DROP TABLE IF EXISTS components;
DROP TYPE IF EXISTS component_status;
