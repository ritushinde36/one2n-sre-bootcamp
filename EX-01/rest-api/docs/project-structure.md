# Project Structure

← [Back to README](../README.md)

This page shows the layout of this repository. It explains what each file and folder does.

```
rest-api/
├── README.md                      # Entry point: quick start plus links to every page below
├── main.go                        # Application entry point: config, database connect, routes, graceful shutdown
├── Makefile                       # build, run, test, migrate-*, docker-*, compose-*, and vagrant-* targets
├── go.mod, go.sum                 # Go module definition and dependency lockfile
├── .env.example                   # Documents all supported environment variables, used both locally and for Docker
├── .gitignore                     # Files and folders excluded from version control
├── Dockerfile                     # Multi-stage build for the app image (see docs/docker.md)
├── docker-entrypoint.sh           # Image entrypoint: execs into rest-api (migrations run separately, see docs/docker.md)
├── docker-compose.yml             # Compose setup: mysql, migrate, and rest-api services (see docs/docker.md)
├── docker-compose.proxy.yml       # Compose setup with nginx in front of two API containers (see docs/vagrant.md)
├── .dockerignore                  # Excludes tests, docs, and env files from the Docker build context
├── Vagrantfile                    # Defines the UTM-backed VM and forwards port 8080 (see docs/vagrant.md)
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
│   ├── minikube.md                          # Running the app on a local Kubernetes cluster
│   ├── postman.md                           # Importing and running the collection, Newman
│   ├── prerequisites.md                     # Tools you need before you start
│   ├── project-structure.md                 # This page: the repo layout, file by file
│   ├── secrets-management.md                # Vault and External Secrets for the Kubernetes deployment
│   ├── setup.md                             # Setting up and running the app on your machine
│   ├── tech-stack.md                        # Libraries and tools, by concern
│   ├── testing.md                           # Running tests and code quality checks
│   ├── troubleshooting.md                   # Common errors and fixes
│   └── vagrant.md                           # Deploying on bare metal in a Vagrant VM
│
├── scripts/
│   ├── install-prerequisites.sh   # Installs everything in Prerequisites if missing (macOS only)
│   └── provision-vm.sh            # Runs inside the VM: installs Docker, starts the proxy stack
│
├── nginx/
│   └── default.conf               # Reverse proxy config: load balancing and the JSON access log
│
├── manifests/                     # Kubernetes deployment (see docs/minikube.md)
│   ├── application.yml            # Namespace, config, external secret, API deployment with migration init container, NodePort service
│   ├── database.yml               # Config, external secret, 2Gi volume claim, MySQL deployment and service
│   ├── secret-store.yml           # ClusterSecretStore pointing the External Secrets Operator at Vault
│   └── values/
│       └── vault-values.yaml      # Helm values for Vault: standalone mode, pinned to the dependent_services node
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
    ├── student-api.postman_collection.json  # Importable Postman collection (see docs/postman.md)
    └── vagrant.postman_environment.json     # Points the collection at the Vagrant VM on port 8080
```
