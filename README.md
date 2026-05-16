# gator

A CLI RSS aggregator written in Go.

## Prerequisites

You need both of these installed before running `gator`:

1. PostgreSQL (the app stores users, feeds, follows, and posts in Postgres)
2. Go (to install/build the CLI)

Quick checks:

```bash
psql --version
go version
```

## Install the CLI

From any directory, install with:

```bash
go install github.com/GravitiMusic/gator@latest
```

If `gator` is not found after install, make sure your Go bin directory is on your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Database setup

Create a Postgres database named `gator` (or any name you prefer):

```bash
createdb gator
```

If you are using `goose` migrations in this repo, run them from the schema folder:

```bash
cd sql/schema
goose postgres "postgres://postgres:postgres@localhost:5432/gator" up
```

Adjust username, password, host, port, and database name to match your local Postgres setup.

## Config file

`gator` reads config from:

`~/.gatorconfig.json`

Create it with this shape:

```json
{
	"db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
	"current_user_name": ""
}
```

Notes:

1. `db_url` must point to your Postgres database.
2. `current_user_name` can be empty at first. It is updated by `register` or `login`.

## Run the program

You can run directly from source:

```bash
go run . <command> [args...]
```

Or use the installed binary:

```bash
gator <command> [args...]
```

## Common commands

```bash
# Create a user and set as current user
gator register alice

# Log in as an existing user
gator login alice

# Add and auto-follow a feed
gator addfeed "Hacker News" "https://news.ycombinator.com/rss"

# Show all feeds
gator feeds

# Follow/unfollow by feed URL
gator follow "https://news.ycombinator.com/rss"
gator unfollow "https://news.ycombinator.com/rss"

# Start background-style feed scraping loop
gator agg 30s

# Browse latest posts for current user
gator browse 5

# Show users (marks current user)
gator users
```

## Command summary

- `register <name>`
- `login <name>`
- `users`
- `reset`
- `addfeed <name> <url>`
- `feeds`
- `follow <url>`
- `following`
- `unfollow <url>`
- `agg <duration>` (example: `30s`, `1m`)
- `browse [limit]`
