-- +goose Up
INSERT INTO components (type, name) VALUES ('orchestration', 'coin-pipeline')
ON CONFLICT DO NOTHING;

INSERT INTO component_versions (component_id, version, content_ref)
SELECT c.id, '1.0.0', '{"artifactKey": "orchestration/coinPipeline.groovy", "sha256": "sha256:c804e7db2d629c40a4e2eafed181784fabc8ad0ab57ee48832561831cca77f6b"}'::jsonb
FROM components c WHERE c.type='orchestration' AND c.name='coin-pipeline'
ON CONFLICT DO NOTHING;

INSERT INTO gp_composition (gp_release_id, component_type, component_name, component_version)
SELECT gr.id, 'orchestration', 'coin-pipeline', '1.0.0'
FROM gp_releases gr
WHERE gr.name='go-app' AND gr.version='1.0.0'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM gp_composition
WHERE gp_release_id IN (SELECT id FROM gp_releases WHERE name='go-app' AND version='1.0.0')
  AND component_type='orchestration';
