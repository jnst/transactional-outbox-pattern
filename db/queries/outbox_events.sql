-- name: CreateOutboxEvent :one
INSERT INTO outbox_events (aggregate_id, event_type, payload) 
VALUES ($1, $2, $3) RETURNING *;

-- name: GetUnpublishedEvents :many
SELECT * FROM outbox_events 
WHERE published_at IS NULL 
ORDER BY created_at ASC 
LIMIT $1;

-- name: MarkEventAsPublished :exec
UPDATE outbox_events 
SET published_at = CURRENT_TIMESTAMP 
WHERE id = $1;