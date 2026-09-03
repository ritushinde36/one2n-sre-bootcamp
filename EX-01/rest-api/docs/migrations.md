# Database Migrations

← [Back to README](../README.md)

This page explains how database migrations work. It covers running them locally, in Docker, and in Compose, plus how to adopt an existing schema.

[goose](https://github.com/pressly/goose) manages schema changes. They live in [migrations/](../migrations/), as paired up and down SQL files.

- `00001_create_students_table.sql`: creates the `students` table
- `00002_add_not_null_constraints.sql`: tightens `name`, `email`, `age`, `class`, and `department` to `NOT NULL`

**Running locally**, against a local MySQL install. Use the Makefile targets in the [Makefile Reference](makefile.md#migrations), or run these directly:

```bash
go run ./cmd/migrate up       # apply all pending migrations
go run ./cmd/migrate down     # roll back the last migration
go run ./cmd/migrate status   # show applied/pending migrations
```

**Running via the manual [Docker workflow](docker.md#running-with-docker):**

```bash
make docker-migrate                     # apply all pending migrations
make docker-migrate MIGRATE_CMD=down    # roll back the last migration
make docker-migrate MIGRATE_CMD=status  # show applied/pending migrations
```

Run `up` once, before `make docker-run`. Migrations do not run automatically every time you run `docker-run`. `cmd/migrate` also locks the database, so running it more than once at the same time stays safe.

**Running via [Docker Compose](docker.md#running-with-docker-compose):**

`make compose-up` already runs `up` for you, before it starts `rest-api`. You do not need to run anything separately for that.

For `down`, `status`, or to rerun `up` on demand, use `compose-migrate`. It runs the command in its own, separate container, safe to run alongside the container that `compose-up` manages:

```bash
make compose-migrate                     # apply all pending migrations
make compose-migrate MIGRATE_CMD=down    # roll back the last migration
make compose-migrate MIGRATE_CMD=status  # show applied/pending migrations
```

**Adopting a pre-existing schema**

A `students` table might already exist, for example if you created it by hand, or with an app version from before goose. If so, and goose has no migration history yet, the migration tool checks the table first. It compares the existing table's schema (`SHOW CREATE TABLE`) against migration `00001`:

- **If they match**, it marks `00001` as already applied. It does not try to recreate the table.
- **If they do not match**, it stops. It prints both schemas, so you can reconcile them by hand. This acts as a safety check.

**Migration logging**

Migration logs use structured JSON, through slog, the same as the app itself. See [cmd/migrate/logger.go](../cmd/migrate/logger.go).
