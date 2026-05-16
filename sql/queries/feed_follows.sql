-- name: CreateFeedFollow :one
WITH feed_follow AS (
    INSERT INTO feed_follows(id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id, created_at, updated_at, user_id, feed_id
)
SELECT feed_follow.id, feed_follow.created_at, feed_follow.updated_at, feed_follow.user_id, feed_follow.feed_id, feeds.name AS feed_name, users.name AS user_name
FROM feed_follow
JOIN feeds ON feed_follow.feed_id = feeds.id
JOIN users ON feed_follow.user_id = users.id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.id, feed_follows.created_at, feed_follows.updated_at, feed_follows.user_id, feed_follows.feed_id, feeds.name AS feed_name, users.name AS user_name
FROM feed_follows
JOIN feeds ON feed_follows.feed_id = feeds.id
JOIN users ON feed_follows.user_id = users.id
WHERE feed_follows.user_id = $1;

-- name: DeleteFeedFollowByUserAndFeed :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;