-- +goose Up
-- Migrate legacy gitRef/path content_ref rows to artifactKey refs (idempotent for fresh seeds).
UPDATE component_versions cv
SET content_ref = '{
    "stages": [
        {"name": "validate", "artifactKey": "scripts/validate.sh", "sha256": "sha256:2be1a1953aa39ac72ac1f13217e15a639060c86451dc89d0dfd30cbef1d8283e"},
        {"name": "test", "artifactKey": "scripts/test.sh", "sha256": "sha256:e406acf01fe536b3778af7359763d3fa891a55036b1e9c5e6e533089f51caf4b"},
        {"name": "build", "artifactKey": "scripts/build.sh", "sha256": "sha256:72f82ac533ce4c3ab7dcd9c72bf86f8f0b0447fe2856ac3b12fac807f6770f25"},
        {"name": "publish", "when": "tag", "artifactKey": "scripts/publish.sh", "sha256": "sha256:0ee3affe70601b98a05c2a8dd1460c147a23af159e13deb40fb2cd377a572948"}
    ]
}'::jsonb
FROM components c
WHERE cv.component_id = c.id
  AND c.type = 'pipeline' AND c.name = 'go-build' AND cv.version = '2.1.0'
  AND cv.content_ref::text LIKE '%gitRef%';

UPDATE component_versions cv
SET content_ref = '{"artifactKey": "schema/config.v2.schema.json", "sha256": "sha256:b13f17b71360985d2e5391490a8b7076718145dd962037ee39573e8d0381c8ab"}'::jsonb
FROM components c
WHERE cv.component_id = c.id
  AND c.type = 'validate' AND c.name = 'config' AND cv.version = '1.0.0'
  AND cv.content_ref::text LIKE '%gitRef%';

UPDATE component_versions cv
SET content_ref = '{"artifactKey": "Dockerfile", "sha256": "sha256:0f3ab76998d2e52d27e019d9173540bf7dfb586577c51cd3b55a7f2ee90d6326"}'::jsonb
FROM components c
WHERE cv.component_id = c.id
  AND c.type = 'dockerfile' AND c.name = 'go-runtime' AND cv.version = '1.0.0'
  AND cv.content_ref::text LIKE '%gitRef%';

-- +goose Down
-- content_ref rollback not supported; re-run legacy seed manually if needed.
