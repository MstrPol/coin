-- +goose Up
-- Normalize legacy published GP versions that incorrectly kept -snapshot.N suffix.

UPDATE catalog_policy cp
SET latest = regexp_replace(cp.latest, '-snapshot\.[0-9]+$', '')
WHERE latest LIKE '%-snapshot.%';

UPDATE catalog_policy cp
SET minimum = regexp_replace(cp.minimum, '-snapshot\.[0-9]+$', '')
WHERE minimum LIKE '%-snapshot.%';

UPDATE catalog_policy cp
SET latest_canary = regexp_replace(cp.latest_canary, '-snapshot\.[0-9]+$', '')
WHERE latest_canary IS NOT NULL AND latest_canary LIKE '%-snapshot.%';

-- Drop snapshot published rows when canonical published already exists (unique name+version).
DELETE FROM gp_releases snap
WHERE snap.status = 'published'
  AND snap.version LIKE '%-snapshot.%'
  AND EXISTS (
    SELECT 1 FROM gp_releases canon
    WHERE canon.name = snap.name
      AND canon.status = 'published'
      AND canon.version = regexp_replace(snap.version, '-snapshot\.[0-9]+$', '')
  );

UPDATE gp_releases
SET version = regexp_replace(version, '-snapshot\.[0-9]+$', '')
WHERE status = 'published' AND version LIKE '%-snapshot.%';

-- +goose Down
-- Irreversible data normalization.
