# Project Structure

← [Back to README](../README.md)

This page shows the layout of this repository. It explains what each file and folder does.

```
rest-api/
├── README.md                      # Entry point: quick start plus links to every page below
├── main.go                        # Application entry point: config, database connect, routes, graceful shutdown
├── Makefile                       # build, run, test, migrate-*, docker-*, and compose-* targets
├── go.mod, go.sum                 # Go module definition and dependency lockfile
├── .env.example                   # Documents all supported environment variables, used both locally and for Docker
├── .gitignore                     # Files and folders excluded from version control
├── Dockerfile                     # Multi-stage build for the app image (see docs/docker.md)
├── docker-entrypoint.sh           # Image entrypoint: execs into rest-api (migrations run separately, see docs/docker.md)
├── docker-compose.yml             # Compose setup: mysql, migrate, and rest-api services (see docs/docker.md)
├── .dockerignore                  # Excludes tests, docs, and env files from the Docker build context
│
├── docs/                          # This documentation, one focused page per topic
│   ├── api-reference.md                     # Endpoints, request/response formats, pagination, status codes
│   ├── architecture.md                      # How the client, REST API, database, and migrations fit together
│   ├── ci-cd.md                             # The GitHub Actions pipeline: steps, triggers, image publishing
│   ├── data-model.md                        # The `Student` schema and its constraints
│   ├── docker.md                            # Running the app with Docker and Docker Compose
│   ├── environment-variables.md             # Every `.env` key, what it does, and its default
│   ├── features.md                          # Full feature list
│   ├── graceful-shutdown.md                 # SIGINT/SIGTERM handling, draining in-flight requests
│   ├── health-and-readiness-checks.md       # Liveness and readiness checks for the app
│   ├── logging.md                           # Log format and sources
│   ├── makefile.md                          # Every `make` target, grouped by what it does
│   ├── migrations.md                        # Running migrations locally, in Docker, and in Compose
│   ├── minikube.md                          # Setting up the local Kubernetes cluster
│   ├── postman.md                           # Importing and running the collection, Newman
│   ├── prerequisites.md                     # Tools you need before you start
│   ├── project-structure.md                 # This page: the repo layout, file by file
│   ├── setup.md                             # Setting up and running the app on your machine
│   ├── tech-stack.md                        # Libraries and tools, by concern
│   ├── testing.md                           # Running tests and code quality checks
│   └── troubleshooting.md                   # Common errors and fixes
│
├── scripts/
│   └── install-prerequisites.sh   # Installs everything in Prerequisites if missing (macOS only)
│
├── config/
│   └── load_config.go             # Loads environment variables from .env via godotenv
│
├── connections/
│   ├── db_connection.go           # Opens the GORM connection to MySQL, exposes the shared `DB` handle
│   └── gorm_logger.go             # Routes GORM's internal logging through slog (structured JSON)
│
├── controllers/
│   └── StudentController.go       # HTTP handlers: CRUD for students, healthcheck, readyz
│
├── models/
│   └── studentModel.go            # `Student` GORM model and database schema struct
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
├── controllers_test/              # Integration tests (see docs/testing.md)
│   ├── main_test.go                    # TestMain: starts one shared MySQL Testcontainer for the whole suite
│   ├── helpers_test.go                 # Shared test router and request helpers
│   ├── StudentController_test.go       # Malformed-input or bad-JSON edge cases for Create & Update
│   ├── student_endpoints_test.go       # Per-endpoint behavior: happy paths, 404s, 409s, validation
│   └── student_lifecycle_test.go       # End-to-end create→read→update→read→delete→read flow
│
└── postman/
    └── student-api.postman_collection.json  # Importable Postman collection (see docs/postman.md)
```
