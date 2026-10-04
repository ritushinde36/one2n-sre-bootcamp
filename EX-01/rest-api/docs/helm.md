# Helm Charts

← [Back to README](../README.md)

This page explains how to deploy the app and its services with the Helm charts in [helm/](../helm/). It covers what each chart installs, the install order, and how to change a setting.

You need `minikube`, `kubectl` and `helm`. See [Prerequisites](prerequisites.md).

## The charts

The Make targets install five charts. Each chart becomes a Helm release, with a fixed release name, in a fixed namespace:

| Chart | Installs | Release name | Namespace |
| --- | --- | --- | --- |
| [external-secrets](../helm/external-secrets/) | The External Secrets Operator | `external-secrets` | `external-secrets` |
| [vault](../helm/vault/) | Vault, in standalone mode | `vault` | `vault` |
| [secret-store](../helm/secret-store/) | The `vault-backend` ClusterSecretStore | `secret-store` | `external-secrets` |
| [mysql](../helm/mysql/) | MySQL | `mysql` | `student-api` |
| [student-api](../helm/student-api/) | The REST API | `student-api` | `student-api` |

### Do not change the release names

Each chart names its objects after the release name. Other parts of the stack look for these objects by name:

| Release | Creates | Used by |
| --- | --- | --- |
| `mysql` | The Service `mysql` in `student-api` | The app. The DSN in Vault connects to `mysql.student-api.svc`. |
| `vault` | The Service `vault` in `vault` | The secret store. It tells the operator to reach Vault at `vault.vault.svc`. |
| `external-secrets` | The service account `external-secrets` in `external-secrets` | Vault. Vault accepts a login only from this service account. |
| `student-api` | The Service `student-api` in `student-api` | `make k8s-port-forward`. |

A different release name gives the object a different name. For example, if you install `helm/mysql` as the release `db`, the Service is `db-mysql`. The app then cannot find the database.

`vault` and `external-secrets` use community charts, from HashiCorp and from the External Secrets project. The repository keeps a copy of each community chart as a `.tgz` file in the `charts/` folder of the chart.

## 1. Create the cluster

```bash
make k8s-cluster-up
```

The command does four things:

1. Creates a cluster of 4 nodes — 1 control plane and 3 workers.
2. Waits until every node reports ready.
3. Labels each worker with its `type`.
4. Prints the nodes and their labels.

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

## 2. Install the operator and Vault

```bash
make helm-dependent-services-up
```

The command does two things:

1. Installs the External Secrets Operator from `helm/external-secrets`, and waits until its pods are ready.
2. Installs Vault from `helm/vault`. It does not wait, because Vault is not ready until you unseal it.

Check the pods:

```bash
kubectl get pods -n external-secrets -o wide
kubectl get pods -n vault
```

Expect this:

- Three operator pods at `1/1 Running`, on `minikube-m04`.
- `vault-0` at `0/1 Running`. This is correct. Vault starts sealed, and a sealed Vault reports that it is not ready.

Check the releases:

```bash
make helm-list
```

Expect two releases at revision `1`, with the status `deployed`:

```
NAME              NAMESPACE         REVISION  UPDATED  STATUS    CHART                   APP VERSION
external-secrets  external-secrets  1         <date>   deployed  external-secrets-0.1.0
vault             vault             1         <date>   deployed  vault-0.1.0
```

- `CHART` shows the version of our wrapper chart, `0.1.0`. The community chart version is in the `Chart.yaml` of the wrapper.
- `APP VERSION` is empty, because the wrapper charts have no `appVersion`.
- Each later run of `make helm-dependent-services-up` adds 1 to `REVISION`, even when nothing changed.

## 3. Configure Vault

Vault starts sealed and empty. Follow [Vault Setup](vault-setup.md). It has three steps:

1. Initialise and unseal Vault.
2. Write the secrets.
3. Configure Kubernetes authentication.

Then return here for step 4.

## 4. Deploy the application

```bash
make helm-deploy
```

The command does five things, in this order:

1. Installs the `secret-store` chart.
2. Creates the `student-api` namespace, if it does not exist.
3. Labels the namespace for Pod Security: `enforce=baseline` and `warn=restricted`.
4. Installs the `mysql` chart, and waits until MySQL is ready.
5. Installs the `student-api` chart, and waits until the API is ready.

The database goes in first, because the readiness check of the API needs the database.

No chart creates the namespace. If one release owned the namespace, `helm uninstall` of that release would delete everything in it, the other release too.

At the end of each install, Helm shows the notes of the chart. The `student-api` notes show the port-forward command.

To watch the pods start, open a second terminal:

```bash
kubectl get pods -n student-api -w
```

Expect the same order as with the manifests:

1. `mysql-0` reaches `1/1 Running`.
2. The `student-api` pod shows `Init:0/1` while migrations run.
3. The `student-api` pod can show `Init:Error` once or twice, if the migration starts before MySQL accepts connections. This is normal. Kubernetes waits, then tries again.
4. The `student-api` pod reaches `1/1 Running`.

Then check the result:

```bash
make helm-status
kubectl get clustersecretstore vault-backend
```

Expect this:

- Both pods at `1/1`.
- Both external secrets at `SecretSynced`.
- Five releases with the status `deployed`.
- The store at `Valid`.

## 5. Reach the API

This step is the same as with the manifests, because both ways name the service `student-api`.

Open the tunnel in its own terminal:

```bash
make k8s-port-forward
```

From another terminal:

```bash
curl http://localhost:8888/healthcheck
make newman
```

`make newman` runs the Postman collection against `localhost:8888`. See [Postman Collection](postman.md).

## Change a setting

Each chart keeps its settings in its `values.yaml`. Each setting has a comment that explains it.

**To change a setting for everyone:**

1. Edit the `values.yaml` of the chart.
2. Increase `version` in the `Chart.yaml` of the chart, for example from `0.1.0` to `0.1.1`.
3. Run `make helm-deploy`. For Vault or the operator, run `make helm-dependent-services-up`.
4. Commit both files.

**To try a value one time**, give it on the command line:

```bash
helm upgrade --install student-api helm/student-api -n student-api --wait --set ginMode=debug
```

The next `make helm-deploy` returns the release to the values in `values.yaml`. The Make targets use `--reset-values` for this. Without that flag, an upgrade that gives no values keeps the values of the last revision.

To see the values that a release uses now:

```bash
helm get values student-api -n student-api --all
```

### Deploy a different build

```bash
helm upgrade --install student-api helm/student-api -n student-api --wait --set image.tag=<tag>
```

`<tag>` is an image tag that CI pushed to GHCR. See [CI/CD](ci-cd.md). The `migrate` init container and the API container use the same tag.

To keep the build, set `image.tag` in [helm/student-api/values.yaml](../helm/student-api/values.yaml) to the tag, and commit it. Without that, the next `make helm-deploy` goes back to the tag in `values.yaml`. CI also sets this tag after every successful build on `main`.

## Day-to-day commands

| Command | What it does |
| --- | --- |
| `make helm-status` | Shows the pods, the external secrets, and every release |
| `make helm-list` | Shows every release, with its revision and status |
| `make k8s-port-forward` | Opens `localhost:8888` onto the API |
| `kubectl logs -n student-api deploy/student-api -f` | Shows the application logs |
| `kubectl logs -n student-api deploy/student-api -c migrate` | Shows the migration output |
| `helm history student-api -n student-api` | Lists every revision of the release |
| `helm rollback student-api <revision> -n student-api` | Returns the release to an earlier revision |
| `helm get manifest student-api -n student-api` | Shows the YAML that Helm applied for the release |
| `make helm-lint` | Checks every chart in `helm/`. Needs no cluster |
| `helm template student-api helm/student-api -n student-api` | Shows the YAML that a chart makes. Needs no cluster |

See [Makefile Reference](makefile.md#helm) for every `helm-*` target.

## Cleanup

Remove the app and the database, and keep the cluster:

```bash
make helm-down
```

This keeps more than `make k8s-down`:

- It keeps the `student-api` namespace.
- It keeps the volume claim `data-mysql-0`, and the MySQL data in it.

The next `make helm-deploy` starts with the old students. Vault, the operator and the store stay installed.

Remove everything, Vault and its secrets included:

```bash
make k8s-cluster-down
```

See [Troubleshooting](troubleshooting.md) for common problems.
