-- name: CreateFeed :one
INSERT INTO feeds(id, created_at, updated_at, name, url, user_id)
VALUES (
           $1,
           $2,
           $3,
           $4,
           $5,
           $6
       )
RETURNING *;

-- name: GetAllFeed :many
SELECT * from feeds;

-- name: GetFeedByUrl :one
SELECT * from feeds WHERE url = $1;

-- name: GetFeed :one
SELECT * from feeds WHERE id = $1;


-- name: GetAllFeedsWithUser :many
SELECT f.id, f.created_at, f.updated_at, f.name, f.url, f.user_id, u.name as user_name
FROM feeds f
INNER JOIN users u ON f.user_id = u.id;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1,
    updated_at = $1
WHERE id = $2;

-- name: GetNextFeedToFetch :one
SELECT id, created_at, updated_at, name, url, user_id, last_fetched_at
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;