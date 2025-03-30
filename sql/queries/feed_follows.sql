
-- name: CreateFeedFollow :one
WITH inserted_feed_follows AS (
    INSERT INTO feed_follows(id, created_at, updated_at, user_id, feed_id)
        VALUES (
                   $1,
                   $2,
                   $3,
                   $4,
                   $5
               )
    RETURNING *
)
SELECT inserted_feed_follows.*, feeds.name AS feed_name, users.name AS user_name
FROM inserted_feed_follows
INNER JOIN feeds ON inserted_feed_follows.feed_id = feeds.id
INNER JOIN users ON inserted_feed_follows.user_id = users.id;

-- name: GetFollowedFeedsForUser :many
SELECT f.id, f.created_at, f.updated_at, f.name, f.url, f.user_id
FROM feeds f
         INNER JOIN feed_follows ff ON f.id = ff.feed_id
WHERE ff.user_id = $1; 