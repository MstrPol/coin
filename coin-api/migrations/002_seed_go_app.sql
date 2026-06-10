-- +goose Up
INSERT INTO components (type, name) VALUES
  ('executor', 'coin-executor'),
  ('agent', 'go'),
  ('pipeline', 'go-build'),
  ('validate', 'config'),
  ('dockerfile', 'go-runtime'),
  ('orchestration', 'coin-pipeline')
ON CONFLICT DO NOTHING;

INSERT INTO component_versions (component_id, version, metadata, content_ref)
SELECT c.id, '0.1.0', '{"url":"http://nexus:8081/repository/coin-executor/0.1.0/coin-executor-linux-amd64"}'::jsonb, NULL
FROM components c WHERE c.type='executor' AND c.name='coin-executor'
ON CONFLICT DO NOTHING;

INSERT INTO component_versions (component_id, version, metadata)
SELECT c.id, '1.22.5', '{"image":"nexus:8082/coin-docker/ci-go:1.22-r2","digest":"sha256:7b1448c77987e590f47990d479571549ee0ca02af2fcf00d0f428a17657b1719"}'::jsonb
FROM components c WHERE c.type='agent' AND c.name='go'
ON CONFLICT DO NOTHING;

INSERT INTO component_versions (component_id, version, content_ref)
SELECT c.id, '2.1.0', '{
    "stages": [
        {"name": "validate", "artifactKey": "scripts/validate.sh", "sha256": "sha256:2be1a1953aa39ac72ac1f13217e15a639060c86451dc89d0dfd30cbef1d8283e"},
        {"name": "test", "artifactKey": "scripts/test.sh", "sha256": "sha256:e406acf01fe536b3778af7359763d3fa891a55036b1e9c5e6e533089f51caf4b"},
        {"name": "build", "artifactKey": "scripts/build.sh", "sha256": "sha256:72f82ac533ce4c3ab7dcd9c72bf86f8f0b0447fe2856ac3b12fac807f6770f25"},
        {"name": "publish", "when": "tag", "artifactKey": "scripts/publish.sh", "sha256": "sha256:0ee3affe70601b98a05c2a8dd1460c147a23af159e13deb40fb2cd377a572948"}
    ]
}'::jsonb
FROM components c WHERE c.type='pipeline' AND c.name='go-build'
ON CONFLICT DO NOTHING;

INSERT INTO component_versions (component_id, version, content_ref)
SELECT c.id, '1.0.0', '{"artifactKey": "schema/config.v2.schema.json", "sha256": "sha256:b13f17b71360985d2e5391490a8b7076718145dd962037ee39573e8d0381c8ab"}'::jsonb
FROM components c WHERE c.type='validate' AND c.name='config'
ON CONFLICT DO NOTHING;

INSERT INTO component_versions (component_id, version, content_ref)
SELECT c.id, '1.0.0', '{"artifactKey": "Dockerfile", "sha256": "sha256:0f3ab76998d2e52d27e019d9173540bf7dfb586577c51cd3b55a7f2ee90d6326"}'::jsonb
FROM components c WHERE c.type='dockerfile' AND c.name='go-runtime'
ON CONFLICT DO NOTHING;

INSERT INTO component_versions (component_id, version, content_ref)
SELECT c.id, '1.0.0', '{"artifactKey": "orchestration/coinPipeline.groovy", "sha256": "sha256:c804e7db2d629c40a4e2eafed181784fabc8ad0ab57ee48832561831cca77f6b"}'::jsonb
FROM components c WHERE c.type='orchestration' AND c.name='coin-pipeline'
ON CONFLICT DO NOTHING;

INSERT INTO gp_releases (name, version, git_export_tag)
VALUES ('go-app', '1.0.0', 'go-app/v1.0.0')
ON CONFLICT DO NOTHING;

INSERT INTO gp_composition (gp_release_id, component_type, component_name, component_version)
SELECT gr.id, x.component_type, x.component_name, x.component_version
FROM gp_releases gr
CROSS JOIN (VALUES
  ('executor', 'coin-executor', '0.1.0'),
  ('agent', 'go', '1.22.5'),
  ('pipeline', 'go-build', '2.1.0'),
  ('validate', 'config', '1.0.0'),
  ('dockerfile', 'go-runtime', '1.0.0'),
  ('orchestration', 'coin-pipeline', '1.0.0')
) AS x(component_type, component_name, component_version)
WHERE gr.name='go-app' AND gr.version='1.0.0'
ON CONFLICT DO NOTHING;

INSERT INTO catalog_policy (gp_name, latest, minimum, deprecated)
VALUES ('go-app', '1.0.0', '1.0.0', '[]'::jsonb)
ON CONFLICT (gp_name) DO UPDATE SET latest=EXCLUDED.latest, minimum=EXCLUDED.minimum;

-- +goose Down
DELETE FROM gp_composition WHERE gp_release_id IN (SELECT id FROM gp_releases WHERE name='go-app' AND version='1.0.0');
DELETE FROM gp_releases WHERE name='go-app' AND version='1.0.0';
DELETE FROM catalog_policy WHERE gp_name='go-app';
