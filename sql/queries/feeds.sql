-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_ID)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetAllFeeds :many
SELECT feeds.name, feeds.url, users.name from feeds
JOIN users ON feeds.user_ID = users.id;