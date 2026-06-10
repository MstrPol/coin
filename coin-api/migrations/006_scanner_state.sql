-- +goose Up
CREATE TABLE scanner_state (
    repo_full_name TEXT PRIMARY KEY,
    last_sha        TEXT NOT NULL,
    last_scan_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS scanner_state;
