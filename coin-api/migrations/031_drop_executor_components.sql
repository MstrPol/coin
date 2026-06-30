-- +goose Up
-- Agent stack is the sole CI runtime component; executor registry rows are obsolete.
DELETE FROM component_versions cv
USING components c
WHERE cv.component_id = c.id AND c.type = 'executor';

DELETE FROM components WHERE type = 'executor';

-- +goose Down
-- Re-seed via bootstrap if rollback needed (seed-jenkins-lib-stack.sh).
