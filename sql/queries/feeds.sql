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

-- name: GetFeedByURL :one
SELECT id, name, url from feeds
WHERE url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
where id = $1;

-- name: GetNextFeedToFetch :one
SELECT * from feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;