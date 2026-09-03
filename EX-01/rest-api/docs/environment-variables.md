# Environment Variables

← [Back to README](../README.md)

This page documents every environment variable the app reads. It covers what each one does, its default, and when it is required. Use it as a reference when you configure `.env`.

Documented in [.env.example](../.env.example):

| Variable | Required | Default | Description |
|---|---|---|---|
| `DSN` | Yes | — | MySQL connection string. |
| `PORT` | No | `8888` | Port the HTTP server listens on. |
| `GIN_MODE` | No | `debug` | Gin's runtime mode: `debug`, `release`, or `test`. |
| `MYSQL_ROOT_PASSWORD` | Only for Docker | — | Root password for the `mysql` container. |
| `MYSQL_DATABASE` | Only for Docker | — | Database the `mysql` container creates. |
| `MYSQL_PORT` | No | `3306` | Host port for MySQL, when running via Docker. |
| `LOG_FILE` | No | stdout only | Where the app also writes its logs. |
| `MIGRATE_LOG_FILE` | No | stdout only | Where `cmd/migrate` also writes its logs. |

**Notes:**

- `DSN`: the app exits immediately if this is unset, or the connection fails. Its host differs by mode: `127.0.0.1` when running directly, `student-mysql` when running via Docker. See [Docker & Docker Compose](docker.md).
- `GIN_MODE`: set to `release` for production. See [main.go](../main.go).
- `MYSQL_ROOT_PASSWORD` and `MYSQL_DATABASE`: read only by the `mysql` container itself, on first boot. The app never reads them.
- `MYSQL_PORT`: applies to both the manual Docker flow and Compose. Override it if something on your machine already listens on 3306. It does not affect `DSN`. Other services always reach MySQL as `student-mysql:3306`.
- `LOG_FILE` and `MIGRATE_LOG_FILE`: set these to persist logs on a volume (`student-logs`, `student-migrate-logs`), so they survive after the container is removed. See [main.go](../main.go) and [cmd/migrate/main.go](../cmd/migrate/main.go).

Note: `.env` loads automatically at startup, and it is gitignored. The same file works for both running the app directly and running it via Docker. See [Docker & Docker Compose](docker.md) for what changes between the two.
