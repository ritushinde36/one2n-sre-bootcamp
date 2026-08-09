# Student REST API

A Go-based REST API for doing CRUD operations on student records, built with Gin, GORM, and MySQL. It includes : structured logging, graceful shutdown, health/readiness probes, versioned DB migrations, and an integration test suite.

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [How It Works](#how-it-works)
- [Data Model](#data-model)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Running the App](#running-the-app)
- [Running with Docker](#running-with-docker)
- [Database Migrations](#database-migrations)
- [API Reference](#api-reference)
- [Postman Collection](#postman-collection)
- [Testing](#testing)
- [Static Analysis](#static-analysis)
- [Logging](#logging)
- [Health &amp; Readiness Checks](#health--readiness-checks)
- [Graceful Shutdown](#graceful-shutdown)
- [Environment Variables](#environment-variables)
- [Troubleshooting](#troubleshooting)

## Features

- Create, read (single/list with pagination), update, and delete student records
- MySQL persistence via GORM, with `TranslateError` enabled so DB-level errors surface as typed Go errors instead of raw driver errors
- Router built on Gin
- Versioned, reversible database schema migrations via [goose](https://github.com/pressly/goose)
- Structured JSON logging for both the app and the migration tool (`log/slog`)
- Liveness (`/healthcheck`) and readiness (`/readyz`) probes, suitable for container orchestrators
- Graceful shutdown on `SIGINT`/`SIGTERM` (drains in-flight requests, 10s timeout)
- Integration test suite that runs against a real, disposable MySQL instance (via Testcontainers)
- Postman collection covering the happy paths as well as the failure cases

## Tech Stack

| Concern | Library |
|---|---|
| HTTP router | [gin-gonic/gin](https://github.com/gin-gonic/gin) |
| ORM | [gorm.io/gorm](https://gorm.io/) + [gorm.io/driver/mysql](https://github.com/go-gorm/mysql) |
| DB driver | [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) |
| Migrations | [pressly/goose](https://github.com/pressly/goose) |
| Config | [joho/godotenv](https://github.com/joho/godotenv) (loads `.env`) |
| Logging | Go standard library `log/slog` (JSON handler) |
| Test assertions | [stretchr/testify](https://github.com/stretchr/testify) |
| Test infrastructure | [testcontainers-go](https://github.com/testcontainers/testcontainers-go) |

Go version: see [go.mod](go.mod) (currently 1.26.5).

## How It Works

On startup ([main.go](main.go)):

1. Structured JSON logging is configured as the default logger.
2. Gin is set to reject unknown JSON fields in request bodies.
3. [`config.LoadConfig()`](config/load_config.go) loads environment variables from a `.env` file in the working directory.
4. [`connections.Connect_to_DB()`](connections/db_connection.go) opens the GORM/MySQL connection using the `DSN` environment variable. If `DSN` is unset or the connection fails, the process logs an error and exits.
5. Routes are registered under `/api/v1` (see [API Reference](#api-reference)), plus `/healthcheck` and `/readyz` at the root.
6. The HTTP server starts in a goroutine; the main goroutine blocks waiting for `SIGINT`/`SIGTERM` to trigger a [graceful shutdown](#graceful-shutdown).

**Request flow for a typical CRUD call:**

```
client → Gin router → SlogLogger middleware → controller (controllers/StudentController.go)
                                                     │
                                                     ├─ binds & validates JSON body into StudentRequest
                                                     ├─ calls connections.DB (GORM) with request context
                                                     ├─ on success → JSON response body + 200/201 status
                                                     └─ on failure → maps GORM errors → HTTP status + JSON error body
```

Requests never bind directly into `models.Student`. Instead, each write goes through `StudentRequest` ([controllers/StudentController.go](controllers/StudentController.go)), a struct that only exposes the client-settable fields (`name`, `email`, `age`, `class`, `department`).

Every DB call is made with `.WithContext(c.Request.Context())`, so if a client disconnects or the request times out, the in-flight query is cancelled instead of continuing to hold a DB connection.

## Data Model

The `Student` struct ([models/studentModel.go](models/studentModel.go)) is both the GORM model and the source of the `students` table schema:

| Field | Type | DB constraints | Client-settable? |
|---|---|---|---|
| `id` | uint | primary key, auto-increment | No — assigned by the DB |
| `created_at` | time.Time | — | No — set automatically by GORM on insert |
| `updated_at` | time.Time | — | No — set automatically by GORM on every update |
| `deleted_at` | time.Time (nullable) | indexed | No — see soft vs. hard delete note below |
| `name` | string | not null, max 255 chars | Yes — required |
| `email` | string | unique, not null, max 255 chars | Yes — required, must be a valid email |
| `age` | int | not null | Yes — required, must be > 0 |
| `class` | string | not null, max 255 chars | Yes — required |
| `department` | string | not null, max 255 chars | Yes — required |

**Soft delete column, but a hard delete:** `deleted_at` comes from GORM's `gorm.DeletedAt` type, which normally enables soft deletes — GORM would set this timestamp instead of removing the row, and exclude soft-deleted rows from queries automatically. However, `DeleteStudent` ([controllers/StudentController.go](controllers/StudentController.go)) explicitly calls `.Unscoped().Delete(...)`, which bypasses that and permanently removes the row. In practice, `deleted_at` is always `null` in this API's responses — deleted students are gone, not soft-deleted.

## Project Structure

```
rest-api/
├── main.go                        # Application entry point: config, DB connect, routes, graceful shutdown
├── Makefile                       # build / run / test / migrate-* targets
├── go.mod / go.sum                # Go module definition and dependency lockfile
├── .env.example                   # Documents all supported environment variables
├── Dockerfile                     # Multi-stage build for the app image (see Running with Docker)
├── .dockerignore                  # Excludes tests, docs, and env files from the Docker build context
│
├── config/
│   └── load_config.go             # Loads environment variables from .env via godotenv
│
├── connections/
│   ├── db_connection.go           # Opens the GORM/MySQL connection, exposes the shared `DB` handle
│   └── gorm_logger.go             # Routes GORM's internal logging through slog (structured JSON)
│
├── controllers/
│   └── StudentController.go       # HTTP handlers: CRUD for students, healthcheck, readyz
│
├── models/
│   └── studentModel.go            # `Student` GORM model / DB schema struct
│
├── middleware/
│   └── logger.go                  # Gin middleware: logs every request as structured JSON
│
├── migrations/
│   ├── 00001_create_students_table.sql   # goose migration: create `students` table
│   └── 00002_add_not_null_constraints.sql # goose migration: tighten column constraints
│
├── cmd/migrate/
│   ├── main.go                    # Standalone CLI: `go run ./cmd/migrate <up|down|status>`
│   └── logger.go                  # Routes goose's internal logging through slog
│
├── controllers_test/              # Integration tests (see Testing section)
│   ├── main_test.go                    # TestMain: starts one shared MySQL Testcontainer for the whole suite
│   ├── helpers_test.go                 # Shared test router/request helpers
│   ├── StudentController_test.go       # Malformed-input / bad-JSON edge cases for Create & Update
│   ├── student_endpoints_test.go       # Per-endpoint behavior: happy paths, 404s, 409s, validation
│   └── student_lifecycle_test.go       # End-to-end create→read→update→read→delete→read flow
│
└── postman/
    └── student-api.postman_collection.json  # Importable Postman collection (see below)
```

## Prerequisites

- [Go](https://go.dev/dl/) (version matching [go.mod](go.mod), currently 1.26.5+)
- A running MySQL server reachable from your machine (local install, or any MySQL 8-compatible instance).
- [Docker](https://www.docker.com/) — **required to run the test suite**, since tests start a real MySQL container via Testcontainers. Also required if you want to run the app itself via containers instead of a local Go toolchain — see [Running with Docker](#running-with-docker).

## Setup


1. **Create the database** in MySQL. Connect to your MySQL server with a username and password (e.g. `mysql -u root -p`), then run:

   ```sql
   CREATE DATABASE student_db;
   ```

   Note the username and password you used to connect — you'll need them again in step 3 to build the `DSN` value.

2. **Install Go dependencies:**

   ```bash
   make deps
   ```

3. **Configure environment variables.** Copy the example file and fill in your MySQL DSN:

   ```bash
   cp .env.example .env
   ```

   See [Environment Variables](#environment-variables) for what each key means. `.env` is loaded automatically on startup by [`config.LoadConfig()`](config/load_config.go); it's gitignored, so your local credentials never get committed.

4. **Run database migrations** to create the schema (see [Database Migrations](#database-migrations) for details):

   ```bash
   make migrate-up
   ```

5. **Start the server:**

   ```bash
   make run
   ```

   The API listens on `http://localhost:8888` by default (override with `PORT`).

## Running the App

The [Makefile](Makefile) defines the standard entry points:

| Command | What it does |
|---|---|
| `make deps` | Installs/tidies Go module dependencies (`go mod tidy`) |
| `make build` | Compiles the binary to `bin/rest-api` |
| `make run` | Builds, then runs `./bin/rest-api` |
| `make test` | Runs the full test suite (`go test ./...`) — **requires Docker running** |
| `make test-coverage` | Runs the full suite with coverage and opens an HTML report — **requires Docker running** |
| `make test-list` | Lists every runnable test's name — no Docker required |
| `make test-one TEST=<name>` | Runs a single test by name — **requires Docker running** |
| `make staticcheck` | Runs static analysis (`staticcheck ./...`) — **requires `staticcheck` installed** |
| `make newman` | Runs the Postman collection against a running server — **requires `newman` installed and the server running** |
| `make migrate-up` | Applies all pending goose migrations |
| `make migrate-down` | Rolls back the most recently applied migration |
| `make migrate-status` | Prints which migrations have been applied |

You can also run directly with `go run .` once `.env` is in place and migrations have been applied.

## Running with Docker

The app can also be built and run as a container, without a local Go toolchain. This uses a separate MySQL container instead of a locally installed MySQL server.

**Files involved:**

- [Dockerfile](Dockerfile) — multi-stage build: compiles both the `rest-api` and `migrate` binaries in a `golang:1.26-alpine` build stage, then copies them (plus `migrations/`) into a minimal `alpine:3.20` runtime image.
- [docker-entrypoint.sh](docker-entrypoint.sh) — the image's entrypoint. Runs `migrate up` to apply any pending migrations, then `exec`s into `rest-api`. Since goose tracks applied migrations in the `goose_db_version` table, this is safe to run on every container start — a container with nothing new to apply just logs `no migrations to run` and moves on.
- [.dockerignore](.dockerignore) — keeps `.env`, `.env.docker`, `bin/`, tests, the Postman collection, and markdown/git files out of the build context.
- `.env.docker` — gitignored env file consumed by the Docker Makefile targets below. It isn't shipped in the repo; create it yourself (step 1).
- [docker-compose.yml](docker-compose.yml) — runs the same app + MySQL setup as one Compose project instead of the manual steps below; see [Running with Docker Compose](#running-with-docker-compose).
- `secrets/` — gitignored directory of credential files (`mysql_root_password.txt`, `dsn.txt`) mounted into containers by Compose's `secrets:` mechanism. Not shipped in the repo; create it yourself (see [Running with Docker Compose](#running-with-docker-compose)).
- [configs/app.env](configs/app.env) — committed, non-secret runtime settings (`PORT`, `GIN_MODE`) mounted into the `rest-api` container by Compose's `configs:` mechanism.

**1. Create `.env.docker`** in the project root:

```
MYSQL_ROOT_PASSWORD=rootpass
MYSQL_DATABASE=student_db
DSN=root:rootpass@tcp(student-mysql:3306)/student_db?charset=utf8mb4&parseTime=True&loc=Local
PORT=8888
```

`student-mysql` is the container name the app connects to over the Docker network created in the next step — that hostname only resolves for containers on that network, not from your host machine.

**2. Start a MySQL container:**

```bash
make docker-mysql-up
```

Creates the `student-api-net` Docker network (if it doesn't already exist) and starts a `mysql:8.0` container named `student-mysql` on it, with data persisted in the `student-mysql-data` volume and port `3306` published to the host.

**3. Build the image:**

```bash
make docker-build
```

Builds `student-rest-api:<version>`, where `<version>` comes from `git describe --tags --always --dirty`.

**4. Run the app container:**

```bash
make docker-run
```

Runs the image detached, named `student-rest-api`, on the `student-api-net` network with `.env.docker` as its environment file, publishing port `8888`. On start, the container's entrypoint applies any pending migrations against the `student-mysql` container before starting the server — no separate migration step needed. The app connects to MySQL using the container-to-container `DSN` from `.env.docker`. Tail its logs with `docker logs -f student-rest-api`.

**5. Verify:**

```bash
curl http://localhost:8888/healthcheck
curl http://localhost:8888/readyz
```

**Cleanup:**

```bash
make docker-down       # gracefully stops (SIGTERM), saves its logs to logs/, and removes the app container
make docker-mysql-down # removes the MySQL container (the student-mysql-data volume persists)
docker network rm student-api-net
```

| Command | What it does |
|---|---|
| `make docker-build` | Builds the app image, tagged with the current git version |
| `make docker-network` | Creates the `student-api-net` Docker network if it doesn't already exist |
| `make docker-mysql-up` | Starts a `mysql:8.0` container on that network, using credentials from `.env.docker` |
| `make docker-run` | Ensures the network exists, then runs the app container detached, using `.env.docker` |
| `make docker-down` | Gracefully stops the app container, saves its logs to `logs/<container>-<timestamp>.log`, then removes it |
| `make docker-mysql-down` | Removes the MySQL container |

Note: [`config.LoadConfig()`](config/load_config.go) only exits on a `.env` read error other than "file not found" — so running in a container with no `.env` file present (which `.dockerignore` guarantees) is fine; environment variables passed via `--env-file` are picked up directly.

### Running with Docker Compose

[docker-compose.yml](docker-compose.yml) replaces the manual network/build/run steps above with two services, `mysql` and `rest-api`. Compose creates its own network per project and attaches both services to it automatically, resolving each by service name (or `container_name`) — there's no equivalent of the `docker-network`/`make docker-network` step to run yourself. The `rest-api` service waits for `mysql`'s healthcheck (`mysqladmin ping`) to pass before starting, since [`Connect_to_DB`](connections/db_connection.go) has no retry/backoff of its own and would otherwise exit if MySQL isn't accepting connections yet.

Credentials and runtime settings are split by sensitivity, using Compose's `secrets:`/`configs:` mechanisms instead of plain env vars for the Compose path (the manual `docker-run`/`docker-mysql-up` targets above are unaffected and still read `.env.docker` directly):

- **Secrets** (`mysql_root_password`, `dsn`) — gitignored files under [secrets/](secrets/), each holding one credential. They're mounted as files at `/run/secrets/<name>` inside the container rather than injected as env vars, so they don't show up in `docker inspect` or a container's process environment. `mysql` consumes its password via the official image's built-in `MYSQL_ROOT_PASSWORD_FILE` support; `rest-api` consumes its `DSN` via `DSN_FILE`, resolved by [`config.GetEnv`](config/load_config.go) (checked in [`connections/db_connection.go`](connections/db_connection.go) and [`cmd/migrate`](cmd/migrate/main.go)) — it reads `KEY_FILE` if set, otherwise falls back to a plain `KEY` env var.
- **Configs** (`app_config` → [configs/app.env](configs/app.env)) — a committed, non-secret settings file (`PORT`, `GIN_MODE`) mounted the same way at `/run/configs/app.env`. Since these aren't sensitive, [docker-entrypoint.sh](docker-entrypoint.sh) just sources the file into the environment before running `migrate`/`rest-api`, which don't have any file-reading convention of their own.

Before first use, create the two secret files under `secrets/` (gitignored, not shipped in the repo):

```bash
mkdir -p secrets
echo -n 'rootpass' > secrets/mysql_root_password.txt
echo -n 'root:rootpass@tcp(student-mysql:3306)/student_db?charset=utf8mb4&parseTime=True&loc=Local' > secrets/dsn.txt
```

With those in place (and `.env.docker` created per step 1 above, which the Compose path still uses for `MYSQL_DATABASE` and the host port mapping):

```bash
make compose-up
```

Builds the app image and starts both containers detached. Verify the same way as above (`curl http://localhost:8888/healthcheck`).

```bash
make compose-down
```

Stops and removes both containers (the `student-mysql-data` volume persists).

| Command | What it does |
|---|---|
| `make compose-up` | Builds the app image and starts `mysql` + `rest-api` detached |
| `make compose-down` | Stops and removes both containers |

## Database Migrations

Schema changes are managed with [goose](https://github.com/pressly/goose) and live in [migrations/](migrations/) as paired up/down SQL files. (This section covers running them yourself against a local MySQL install — if you're using the [Docker workflow](#running-with-docker), `make docker-run` applies pending migrations automatically on container start.)

- `00001_create_students_table.sql` — creates the `students` table
- `00002_add_not_null_constraints.sql` — tightens `name`/`email`/`age`/`class`/`department` to `NOT NULL`

Run them via the Makefile targets above, or directly:

```bash
go run ./cmd/migrate up       # apply all pending migrations
go run ./cmd/migrate down     # roll back the last migration
go run ./cmd/migrate status   # show applied/pending migrations
```

**Adopting a pre-existing schema:** if a `students` table already exists (e.g. it was created manually, or by an app version that predates goose) and goose has no migration history yet, `cmd/migrate` ([cmd/migrate/main.go](cmd/migrate/main.go)) compares the existing table's `SHOW CREATE TABLE` output against migration `00001`. If they match, it marks `00001` as already applied instead of trying to recreate the table. If they don't match, it refuses to proceed and prints both schemas so you can reconcile them manually which acts as a safety check.

Migration logging goes through slog as structured JSON, same as the app itself (see [cmd/migrate/logger.go](cmd/migrate/logger.go)).

## API Reference

Base path for student resources: `/api/v1`

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/students` | List students (paginated) |
| GET | `/api/v1/student/:id` | Get a single student by ID |
| POST | `/api/v1/students` | Create a new student |
| PUT | `/api/v1/student/:id` | Update a student by ID |
| DELETE | `/api/v1/student/:id` | Delete a student by ID |
| GET | `/healthcheck` | Liveness probe (is the process itself running) |
| GET | `/readyz` | Readiness probe (can the app reach its database) |

### Pagination

`GET /api/v1/students` accepts two optional query parameters:

| Param | Meaning | Default |
|---|---|---|
| `limit` | Max number of students to return in one response | 50 |
| `offset` | How many students to skip from the start of the full list, before applying `limit` | 0 |

For example, `limit=10&offset=20` skips the first 20 students and returns the next 10 (students 21–30). This is how you'd fetch "page 3" if each page has 10 students:

```bash
curl "http://localhost:8888/api/v1/students?limit=10&offset=20"
```

### Request body (Create / Update)

Both `POST /api/v1/students` and `PUT /api/v1/student/:id` take the same JSON shape. All fields are required, and **unknown fields are rejected with 400**:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 20,
  "class": "10th",
  "department": "CS"
}
```

| Field | Type | Validation |
|---|---|---|
| `name` | string | required |
| `email` | string | required, must be a valid email, must be unique across students |
| `age` | int | required, must be > 0 |
| `class` | string | required |
| `department` | string | required |

Errors are always returned as `{"error": "<message>"}`.

### Status codes

| Status | When |
|---|---|
| 200 OK | Successful GET / PUT / DELETE |
| 201 Created | Successful POST |
| 400 Bad Request | Malformed JSON, missing/invalid required fields, unknown fields in the body, non-numeric `:id` |
| 404 Not Found | No student exists with the given `:id` |
| 409 Conflict | `email` collides with an existing student |
| 500 Internal Server Error | Unexpected DB failure |
| 503 Service Unavailable | `/readyz` only. Database is unreachable |

### Example requests

The student ID `1` used below is just an example and it's hardcoded for illustration only. Replace it with the actual numeric ID of the student you want to look up, update, or delete.

**Get all students**

```bash
curl http://localhost:8888/api/v1/students
```

**Get one student**

```bash
curl http://localhost:8888/api/v1/student/1
```

**Create a student**

```bash
curl -X POST http://localhost:8888/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 20,
    "class": "10th",
    "department": "CS"
  }'
```

**Update a student**

```bash
curl -X PUT http://localhost:8888/api/v1/student/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 21,
    "class": "11th",
    "department": "CS"
  }'
```

**Delete a student**

```bash
curl -X DELETE http://localhost:8888/api/v1/student/1
```

**Health / readiness**

```bash
curl http://localhost:8888/healthcheck
curl http://localhost:8888/readyz
```

## Postman Collection

A ready-to-import collection lives at [postman/student-api.postman_collection.json](postman/student-api.postman_collection.json). It covers both happy paths (health check, create/get/update/delete) and the main failure cases (missing required fields, not-found lookups, invalid IDs).

**Import into Postman:**

1. Open Postman → **Import** → select `postman/student-api.postman_collection.json`.
2. The collection uses a `base_url` variable (defaults to `http://localhost:8888`).
3. Requests are designed to run **top-to-bottom**: "Create Student" captures the new student's ID into a `student_id` collection variable, which later requests (Get/Update/Delete by ID) reuse. Running requests out of order or in isolation may cause the ID-dependent ones to fail.

**Run headlessly with Newman** (useful in CI or without the Postman GUI). Install it once with `npm install -g newman`, then:

```bash
make newman
# or directly
newman run postman/student-api.postman_collection.json
```

Make sure the server is running (`make run`) before running the collection, whether via the Postman GUI or Newman.

## Testing

```bash
make test
# or
go test ./...
```

**Docker must be running**. `TestMain` in [controllers_test/main_test.go](controllers_test/main_test.go) starts a single real MySQL 8 container (via Testcontainers) shared across every test in the package, and tears the container down after the suite finishes.

### Coverage

```bash
make test-coverage
```

Runs the suite with `-coverprofile`, then opens an HTML report in your browser showing exactly which lines are (and aren't) exercised by a test. Run this locally whenever you want to check.

This uses `-coverpkg=./controllers/...` rather than plain `go test -cover ./...`. That's because the tests live in the separate `controllers_test` package, not in `controllers` itself.

Test files, by what they cover:

| File | Covers |
|---|---|
| [helpers_test.go](controllers_test/helpers_test.go) | Shared router setup and HTTP request helper used by every other test file |
| [StudentController_test.go](controllers_test/StudentController_test.go) | Malformed/invalid JSON bodies for Create and Update (unclosed JSON, wrong types, non-object payloads) |
| [student_endpoints_test.go](controllers_test/student_endpoints_test.go) | Per-endpoint behavior: health/ready checks, listing, not-found (404), duplicate email (409), missing required fields, rejection of unknown fields |
| [student_lifecycle_test.go](controllers_test/student_lifecycle_test.go) | Full create → read → update → read → delete → read flow against the real DB |

Not sure of a test's exact name? List every runnable test (no Docker required):

```bash
make test-list
```

Then run a single test by name:

```bash
make test-one TEST=TestStudentLifecycle_PersistsAcrossRealDB
# or directly
go test ./controllers_test/... -run TestStudentLifecycle_PersistsAcrossRealDB -v
```

## Static Analysis

[Staticcheck](https://staticcheck.dev/) catches issues `go vet` misses like any unused code, suspicious type conversions, deprecated API usage, simplifiable expressions, etc.

**Install it once:**

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

**Run it:**

```bash
make staticcheck
# or directly
staticcheck ./...
```


## Logging

All logging (app and migration tool) goes through `log/slog` with a JSON handler, so every log line is a structured, machine-parseable JSON object rather than free text.

- [middleware/logger.go](middleware/logger.go) logs one line per HTTP request: method, path, status, latency (ms), client IP.
- [connections/gorm_logger.go](connections/gorm_logger.go) routes GORM's own query logging through slog.
- [cmd/migrate/logger.go](cmd/migrate/logger.go) routes goose's own logging through slog.
- Controllers log at `info` for successful operations, `warn` for client errors (bad input, not found), and `error` for unexpected failures (DB errors).

## Health & Readiness Checks

The app exposes two distinct probes, matching the liveness/readiness split used by container orchestrators like Kubernetes:

- **`GET /healthcheck`** — liveness. Confirms only that the process is alive and can respond to HTTP; checks no external dependency. A failure here means the process itself is broken and should be restarted.
- **`GET /readyz`** — readiness. Confirms the app can currently serve real traffic, including pinging the database (with a 2s timeout). A failure here means traffic should stop being routed here — but the process itself doesn't need restarting, since restarting won't fix a database that's down.

## Graceful Shutdown

On `SIGINT` or `SIGTERM`, [main.go](main.go) stops accepting new connections and gives in-flight requests up to 10 seconds to finish (`http.Server.Shutdown` with a context timeout) before exiting. This matters for zero-downtime deploys and container restarts which would kill the process outright and drop any request that was mid-flight.

## Environment Variables

Documented in [.env.example](.env.example):

| Variable | Required | Default | Description |
|---|---|---|---|
| `DSN` | Yes | — | MySQL connection string, e.g. `root:yourpassword@tcp(127.0.0.1:3306)/student_db?charset=utf8mb4&parseTime=True&loc=Local`. The app exits immediately if this is unset or the connection fails. |
| `PORT` | No | `8888` | Port the HTTP server listens on. |

NOTE - `.env` is loaded automatically at startup and is gitignored.

## Troubleshooting

- **App exits immediately with "environment variable DSN is not set"** — you haven't created `.env`. Run `cp .env.example .env` and fill in a real DSN.
- **App exits with "failed to connect to database"** — MySQL isn't running, the DSN host/port/credentials are wrong, or the database named in the DSN doesn't exist yet (see [Setup](#setup) step 1).
- **`make test` hangs or fails to start** — Docker isn't running. The integration suite needs Docker to launch its MySQL Testcontainer.
- **`/readyz` returns 503** — the app is up but can't reach the database; check MySQL is running and reachable from wherever the app is deployed.
- **Migration `up` fails with "refusing to proceed: pre-existing students table does not match migration 00001"** — a `students` table already exists with a schema that doesn't match what migration `00001` expects. Compare the printed `existing` vs `expected` DDL and reconcile manually; `cmd/migrate` will not auto-alter a mismatched table for you.
- **Creating/updating a student returns 400 mentioning an unexpected field name** (e.g. `id`, `created_at`) — the request body included a field the API doesn't allow clients to set. Only `name`, `email`, `age`, `class`, and `department` are accepted.
