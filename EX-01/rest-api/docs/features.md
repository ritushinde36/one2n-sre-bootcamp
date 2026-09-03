# Features

← [Back to README](../README.md)

This page lists what the API can do. It covers the main features, from CRUD operations to testing and CI/CD. Read this for a quick overview of the project.

- Create, read (single or list with pagination), update, and delete student records.
- MySQL persistence through GORM.
- Router built on Gin.
- Versioned, reversible schema migrations, managed by [goose](https://github.com/pressly/goose). A database lock keeps concurrent migration runs safe, and the tool can adopt an existing schema.
- Structured JSON logging for the app and the migration tool.
- Liveness (`/healthcheck`) and readiness (`/readyz`) probes.
- Graceful shutdown on `SIGINT` or `SIGTERM`, with a 10-second timeout to drain in-flight requests.
- An integration test suite. It runs against a real, disposable MySQL instance through Testcontainers. It also reports test coverage.
- Static analysis and linting: `staticcheck` for Go, `hadolint` for the Dockerfile. Both run locally and in CI.
- A Postman collection. It covers the happy paths and the failure cases. Newman can run it headlessly.
- `install-prerequisites.sh` for quick local tooling setup.
- Three ways to run the app: locally, with Docker alone, or with Docker Compose.
- Configurable entirely through environment variables.
- The app writes container logs to a Docker volume. The logs stay readable after the container is gone.
- A CI/CD pipeline. It builds, tests, lints, and publishes a Docker image.
