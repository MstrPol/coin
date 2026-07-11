-- +goose Up
DROP TABLE IF EXISTS platform_settings;

-- +goose Down
CREATE TABLE platform_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    nexus_maven_base TEXT NOT NULL DEFAULT 'http://nexus:8081/repository',
    nexus_credentials_id TEXT NOT NULL DEFAULT 'nexus-admin',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO platform_settings (id) VALUES (1);
