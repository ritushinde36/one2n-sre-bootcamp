# Student REST API

A Golang-based REST API for managing student records, with built-in testing, containerization, and CI/CD.

----

## Getting Started

1. Install the required tools. See [Prerequisites](docs/prerequisites.md).
2. Set up and run the app:
   - Directly on your machine. See [Local Setup](docs/setup.md).
   - In a container. See [Docker & Docker Compose](docs/docker.md).
   - In a Vagrant VM. See [Vagrant VM](docs/vagrant.md).
   - On a Kubernetes cluster, with manifests. See [Minikube Cluster](docs/minikube.md).
   - On a Kubernetes cluster, with Helm charts. See [Helm Charts](docs/helm.md).
   - On a Kubernetes cluster, with Argo CD. See [Argo CD](docs/argocd.md).
3. Explore the available endpoints. See [API Reference](docs/api-reference.md).

----

## Documentation

Everything below is reference material for this project, grouped by when you would need it.

### Understand the Project

| | |
|---|---|
| [Features](docs/features.md) | Full feature list |
| [Tech Stack](docs/tech-stack.md) | Libraries and tools, by concern |
| [Architecture](docs/architecture.md) | How the client, REST API, database, and migrations fit together |
| [Data Model](docs/data-model.md) | The `Student` schema and its constraints |
| [Project Structure](docs/project-structure.md) | The repo layout, file by file |

### Set Up & Run

| | |
|---|---|
| [Prerequisites](docs/prerequisites.md) | Tools you need before you start |
| [Local Setup](docs/setup.md) | Setting up and running the app on your machine |
| [Docker & Docker Compose](docs/docker.md) | Running the app in containers |
| [Vagrant VM](docs/vagrant.md) | Deploying on bare metal in a Vagrant VM |
| [Minikube Cluster](docs/minikube.md) | Running the app on a local Kubernetes cluster |
| [Helm Charts](docs/helm.md) | Running the app on Kubernetes with Helm charts |
| [Argo CD](docs/argocd.md) | Deploying the app on Kubernetes with Argo CD, from the Helm charts in git |
| [Secrets Management](docs/secrets-management.md) | Vault and External Secrets for the Kubernetes deployment |
| [Vault Setup](docs/vault-setup.md) | Initialising, unsealing and configuring Vault, for every Kubernetes way |
| [Environment Variables](docs/environment-variables.md) | Every `.env` key, what it does, and its default |
| [Makefile Reference](docs/makefile.md) | Every `make` target, grouped by what it does |
| [Database Migrations](docs/migrations.md) | Running migrations locally and in Docker |

### Use & Maintain

| | |
|---|---|
| [API Reference](docs/api-reference.md) | Endpoints, pagination, request and response shapes |
| [Postman Collection](docs/postman.md) | Importing and running the collection, Newman |
| [Logging](docs/logging.md) | Log format and sources |
| [Health & Readiness Checks](docs/health-and-readiness-checks.md) | Liveness and readiness checks for the app |
| [Graceful Shutdown](docs/graceful-shutdown.md) | App shutdown and draining in-flight requests |
| [Testing & Static Analysis](docs/testing.md) | Running tests and code quality checks |
| [CI/CD](docs/ci-cd.md) | The GitHub Actions pipeline |
| [Troubleshooting](docs/troubleshooting.md) | Common errors and fixes |
