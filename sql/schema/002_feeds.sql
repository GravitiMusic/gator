-- +goose Up
CREATE TABLE feeds(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    user_ID UUID NOT NULL,
    UNIQUE(url),
    CONSTRAINT fk_user FOREIGN KEY (user_ID) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;