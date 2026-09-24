# Makefile Reference

← [Back to README](../README.md)

The [Makefile](../Makefile) defines every standard entry point for building, running, testing, linting, migrating, and containerizing the app, and for running it in a VM. This page lists every target, grouped by what it does.

## Build & Run

| Command | What it does |
|---|---|
| `make deps` | Installs and updates Go module dependencies |
| `make build` | Builds the binary at `bin/rest-api` |
| `make run` | Builds the binary, then runs it |

## Testing

| Command | What it does |
|---|---|
| `make test` | Runs the full test suite in verbose mode (requires Docker running) |
| `make test-coverage` | Runs the full suite with coverage and opens an HTML report (requires Docker running) |
| `make test-list` | Lists every runnable test's name (no Docker required) |
| `make test-one TEST=<name>` | Runs a single test by name (requires Docker running) |

## Linting & Static Analysis

| Command | What it does |
|---|---|
| `make vet` | Flags buggy code patterns |
| `make fmt` | Checks the code's formatting |
| `make staticcheck` | Runs static analysis (requires `staticcheck` installed) |
| `make hadolint` | Lints the Dockerfile (requires `hadolint` installed) |

## Postman

| Command | What it does |
|---|---|
| `make newman` | Runs the Postman collection against a running server (requires `newman` and the server running) |
| `make newman-vagrant` | Runs the Postman collection against the Vagrant VM on port 8080 (requires `newman` and the VM running) |

## Migrations

For a local MySQL install. Docker and Compose have their own migrate commands below.

| Command | What it does |
|---|---|
| `make migrate-up` | Applies all pending goose migrations |
| `make migrate-down` | Rolls back the most recently applied migration |
| `make migrate-status` | Prints which migrations were applied |

## Docker

| Command | What it does |
|---|---|
| `make docker-build` | Builds the app image, tagged with the current git version |
| `make docker-network` | Creates the Docker network if it does not already exist |
| `make docker-mysql-up` | Starts a MySQL container on that network, with credentials from `.env` |
| `make docker-migrate` | Runs pending migrations to completion in a one-off container |
| `make docker-run` | Starts the app container in detached mode, with configuration from `.env` |
| `make docker-down` | Stops the app container gracefully, then removes it |
| `make docker-mysql-down` | Removes the MySQL container |

## Docker Compose

| Command | What it does |
|---|---|
| `make compose-up` | Builds the app image, starts `mysql`, runs `migrate` to completion, then starts `rest-api` |
| `make compose-migrate` | Runs a one-off migrate command in its own container |
| `make compose-down` | Stops and removes all three containers |

The `compose-proxy-*` targets use [docker-compose.proxy.yml](../docker-compose.proxy.yml) instead. This file adds nginx in front of two API containers. The Vagrant VM runs this same stack inside itself.

| Command | What it does |
|---|---|
| `make compose-proxy-up` | Builds the app image, then starts `mysql`, `migrate`, `api-1`, `api-2`, and `nginx` |
| `make compose-proxy-migrate` | Runs a one-off migrate command against the proxy stack |
| `make compose-proxy-down` | Stops and removes all five proxy-stack containers |

## Vagrant

These targets control the VM from your Mac. See [Vagrant VM](vagrant.md) for the full workflow.

| Command | What it does |
|---|---|
| `make vagrant-up` | Creates the VM, provisions it, and starts the proxy stack inside it |
| `make vagrant-provision` | Re-runs the provisioning script against the existing VM |
| `make vagrant-status` | Reports whether the VM is running, stopped, or not created yet |
| `make vagrant-ssh` | Opens a shell in the VM |
| `make vagrant-halt` | Stops the VM but keeps its disk, so the next `vagrant-up` is fast |
| `make vagrant-destroy` | Deletes the VM and its disk, after a confirmation prompt |

## Kubernetes

These targets run against a local minikube cluster. See [Minikube Cluster](minikube.md) for the full workflow, and [Secrets Management](secrets-management.md) for the Vault setup that `k8s-deploy` depends on.

| Command | What it does |
|---|---|
| `make k8s-cluster-up` | Creates the 4-node cluster and labels each worker by role |
| `make k8s-deploy` | Applies the database and application manifests |
| `make k8s-status` | Shows the pods and whether the external secrets synced |
| `make k8s-port-forward` | Opens `localhost:8888` onto the API (blocks — run in its own terminal) |
| `make k8s-logs` | Tails the API logs (`MIGRATE_LOGS=1` for the migration output) |
| `make k8s-down` | Removes the app and database, keeps the cluster |
| `make k8s-cluster-down` | Deletes the whole cluster |

## Overridable Variables

You can override some target variables on the command line, as `make <target> VARIABLE=value`.

| Variable | Default | Used by | Example |
|---|---|---|---|
| `VERSION` | current git version | `docker-build`, `docker-run` | `make docker-build VERSION=1.2.3` |
| `IMAGE_NAME` | `student-rest-api` | `docker-build`, `docker-run` | `make docker-build IMAGE_NAME=my-student-api` |
| `HOST_PORT` | `8888` | `docker-run` | `make docker-run HOST_PORT=8882` |
| `APP_CONTAINER` | `student-rest-api` | `docker-run`, `docker-down` | `make docker-run APP_CONTAINER=student-rest-api02 HOST_PORT=8882` |
| `MYSQL_PORT` | `3306` | `docker-mysql-up` | `make docker-mysql-up MYSQL_PORT=3307` |
| `MIGRATE_CMD` | `up` | `docker-migrate`, `compose-migrate`, `compose-proxy-migrate` | `make docker-migrate MIGRATE_CMD=status` |
| `K8S_NAMESPACE` | `student-api` | every `k8s-*` target | `make k8s-status K8S_NAMESPACE=other` |
| `MIGRATE_LOGS` | unset | `k8s-logs` | `make k8s-logs MIGRATE_LOGS=1` |

`compose-up` reads its own overridable variables (`PORT`, `MYSQL_PORT`, `IMAGE_NAME`, `VERSION`) from `.env` or the command line. See [Docker & Docker Compose](docker.md) for details.
