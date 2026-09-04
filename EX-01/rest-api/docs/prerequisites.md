# Prerequisites

← [Back to README](../README.md)

This page lists what you need installed before you use this project. It covers required and optional tools, and how to install them.

| Tool | Required? | Notes |
|---|---|---|
| [Go](https://go.dev/dl/) | Yes | Version matching [go.mod](../go.mod), currently 1.26.5+ |
| MySQL | Yes | Reachable from your machine: local install, or any MySQL 8-compatible instance |
| [Docker](https://www.docker.com/) | Yes | Needed for the test suite (Testcontainers), and for running the app in containers. See [Docker & Docker Compose](docker.md). |
| `make` | Yes | Needed to run any Makefile target |
| `git` | Yes | Needed to tag Docker images by version |
| `staticcheck` | Optional | Only for `make staticcheck` |
| `newman` | Optional | Only for `make newman`. Also needs Node.js and npm. |
| `hadolint` | Optional | Only for `make hadolint` |
| [minikube](https://minikube.sigs.k8s.io/) | Yes, for Kubernetes | Needed to run a local cluster. See [Minikube Cluster](minikube.md). |
| `kubectl` | Yes, for Kubernetes | Needed to interact with the cluster. See [Minikube Cluster](minikube.md). |

On macOS, install everything above (skipping anything already present) with:

```bash
./scripts/install-prerequisites.sh
```

See [scripts/install-prerequisites.sh](../scripts/install-prerequisites.sh) to see what it installs, and how. For other platforms, it also lists manual install commands for each tool.

Note: the script installs Docker Desktop, but does not start it.

After installing, open Docker.app once by hand. Check that the menu bar icon shows it running. Do this before using any `docker-*`, `compose-*`, or `test*` target. Otherwise, you will see the error "Cannot connect to the Docker daemon".
