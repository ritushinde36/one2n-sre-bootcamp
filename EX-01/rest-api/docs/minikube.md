# Minikube Cluster

← [Back to README](../README.md)

This page explains how to run the app on a local Kubernetes cluster. It covers creating a multi-node cluster, labelling the nodes by role, and deploying the application to it.

You need `minikube`, `kubectl` and `helm`. See [Prerequisites](prerequisites.md).

## Architecture

```
                 minikube cluster
    ┌──────────────────────────────────────────────┐
    │                                              │
    │  minikube-m02          minikube-m03          │
    │  type=application      type=database         │
    │  ┌────────────────┐    ┌──────────────────┐  │
    │  │ rest-api pod   │───▶│ mysql pod        │  │
    │  │ + migrate init │    │ + 2Gi volume     │  │
    │  └────────────────┘    └──────────────────┘  │
    │         ▲                                    │
    │         │              minikube-m04          │
    │  NodePort service      type=dependent_services│
    │         │              ┌──────────────────┐  │
    │         │              │ Vault            │  │
    │         │              │ External Secrets │  │
    │         │              └──────────────────┘  │
    └─────────┼────────────────────────────────────┘
              │
        your machine
```

Each workload is pinned to a node by label. The API and the database never share a node with each other or with Vault.

## 1. Create the cluster

```bash
make k8s-cluster-up
```

The command does four things:

1. Creates a cluster of 4 nodes — 1 control plane and 3 workers.
2. Waits until every node reports ready.
3. Labels each worker with its `type`.
4. Prints the nodes and their labels.

The first run takes several minutes, because it downloads the node image. Later runs start the existing cluster instead, and are quick.

Each worker gets one label. Every workload uses it to choose a node:

| Node | Label | Runs |
| --- | --- | --- |
| `minikube-m02` | `type=application` | The REST API, and the migration init container |
| `minikube-m03` | `type=database` | MySQL and its persistent volume |
| `minikube-m04` | `type=dependent_services` | Vault and the External Secrets Operator |

The command ends by printing this, so check all four nodes are `Ready` and three carry a `TYPE`:

```
NAME           STATUS   ROLES           VERSION   TYPE
minikube       Ready    control-plane   v1.35.1
minikube-m02   Ready    <none>          v1.35.1   application
minikube-m03   Ready    <none>          v1.35.1   database
minikube-m04   Ready    <none>          v1.35.1   dependent_services
```


## 2. Set up secrets

The manifests hold no secret values. Vault holds them, and the External Secrets Operator syncs them into the cluster.

Follow [Secrets Management](secrets-management.md) before you go further. It installs both tools, initialises Vault, and writes the two secrets the app needs.

Nothing below works until `kubectl get clustersecretstore vault-backend` reports `Valid`.

## 3. Deploy the application

```bash
make k8s-deploy
```

The command applies the database first, then the application. Order matters, because the API waits on the database, and both wait on their secrets.

Watch the pods start:

```bash
kubectl get pods -n student-api -w
```

Expect this order:

1. `mysql` reaches `1/1 Running`.
2. `rest-api` shows `Init:0/1` while migrations run.
3. `rest-api` may show `Init:Error` once or twice, if the migration starts before MySQL accepts connections. This is normal. Kubernetes waits, then tries again.
4. `rest-api` reaches `1/1 Running`.

Then check the result:

```bash
make k8s-status
```

Both pods read `1/1`, and both external secrets read `SecretSynced`.

What each manifest creates:

| File | Creates |
| --- | --- |
| [database.yml](../manifests/database.yml) | Namespace, config, external secret, a 2Gi volume claim, the MySQL deployment, and its service |
| [application.yml](../manifests/application.yml) | Config, external secret, the API deployment with its migration init container, and a NodePort service |

**Migrations run as an init container.** The API container cannot start until that init container has applied every migration and exited successfully. A failed migration leaves the pod in `Init:Error` rather than serving traffic against a half-migrated schema. See [Database Migrations](migrations.md).

**The probes differ on purpose.** Liveness calls `/healthcheck`, which does not touch the database — restarting the API cannot fix a downed database. Readiness calls `/readyz`, which does check the database, so a pod that cannot reach MySQL stops receiving traffic without being restarted. See [Health & Readiness Checks](health-and-readiness-checks.md).

## 4. Reach the API

The service is a NodePort, but on macOS with the `docker` driver the node IP is not routable from the host. A direct call to the node IP and port times out. You need a tunnel.

Open one in its own terminal:

```bash
make k8s-port-forward
```

That maps `localhost:8888` onto the service and **stays in the foreground**. The tunnel closes when you stop the command.

From another terminal:

```bash
curl http://localhost:8888/healthcheck
curl http://localhost:8888/api/v1/students
```

`minikube service student-api -n student-api --url` also works and prints the URL, but it holds its own tunnel open the same way, so it never returns to the prompt.

See [Postman Collection](postman.md) for running the full collection against `http://localhost:8888`.

## The application image

The manifests pull a published image rather than building one:

```
ghcr.io/ritushinde36/student-rest-api:v0.4.5-6-gf42a581
```

The CI pipeline builds and pushes it on every merge, tagged with the Git version. See [CI/CD](ci-cd.md).

To deploy a different build, change the tag in [application.yml](../manifests/application.yml) — it appears twice, once for the init container and once for the API container. Both must match, or migrations run on a different version from the app.

## Day-to-day commands

| Command | What it does |
| --- | --- |
| `make k8s-status` | The pods, and whether the secrets synced |
| `make k8s-logs` | Application logs |
| `make k8s-logs MIGRATE_LOGS=1` | Migration output |
| `make k8s-port-forward` | Open `localhost:8888` onto the API |
| `kubectl describe pod -n student-api <pod>` | Why a pod is not starting |
| `kubectl rollout restart deployment rest-api -n student-api` | Restart the API |

The last two have no make target, because both need you to choose a pod or read the output carefully. See [Makefile Reference](makefile.md#kubernetes) for the full list.

## Cleanup

Remove the application, keep the cluster:

```bash
make k8s-down
```

This deletes the namespace, and the volume claim with it, so the database data is lost.

Remove everything, Vault and its secrets included:

```bash
make k8s-cluster-down
```

See [Troubleshooting](troubleshooting.md) for common problems.
