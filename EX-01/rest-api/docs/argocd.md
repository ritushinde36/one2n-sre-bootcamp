# Argo CD

← [Back to README](../README.md)

This page explains how to deploy the app and its services with Argo CD. Argo CD reads the Helm charts in [helm/](../helm/) from GitHub, and keeps the cluster in sync with them.

You need `minikube`, `kubectl` and `helm`. The `argocd` CLI is optional. See [Prerequisites](prerequisites.md).

## The apps

`make argocd-apps` applies only one app: the root app, in [argocd/root.yaml](../argocd/root.yaml). Its source is the folder [argocd/apps/](../argocd/apps/). Argo CD creates every object in that folder, and keeps it in sync with git:

| Object | Sync wave | Deploys | Namespace |
| --- | --- | --- | --- |
| AppProject `student-api` | -1 | Allows only this repository, and the three namespaces below | — |
| Repository Secret | -1 | Registers this repository with Argo CD. It holds only the URL, because the repository is public | — |
| App `external-secrets` | 0 | [helm/external-secrets](../helm/external-secrets/) | `external-secrets` |
| App `vault` | 0 | [helm/vault](../helm/vault/) | `vault` |
| App `secret-store` | 1 | [helm/secret-store](../helm/secret-store/) | `external-secrets` |
| App `mysql` | 2 | [helm/mysql](../helm/mysql/) | `student-api` |
| App `student-api` | 3 | [helm/student-api](../helm/student-api/) | `student-api` |

- Argo CD syncs the waves in order. It starts a wave only when every app in the earlier waves is Healthy.
- Every app syncs by itself (automated sync), deletes what you remove from git (prune), and undoes changes made outside git (self-heal).

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
| `minikube-m04` | `type=dependent_services` | Argo CD, Vault and the External Secrets Operator |

The command ends by printing this, so check all four nodes are `Ready` and three carry a `TYPE`:

```
NAME           STATUS   ROLES           VERSION   TYPE
minikube       Ready    control-plane   v1.35.1
minikube-m02   Ready    <none>          v1.35.1   application
minikube-m03   Ready    <none>          v1.35.1   database
minikube-m04   Ready    <none>          v1.35.1   dependent_services
```

## 2. Install Argo CD

```bash
make argocd-install
```

The command installs Argo CD from [helm/argocd](../helm/argocd/) into the `argocd` namespace. It waits until all Argo CD pods are ready.

Check the pods:

```bash
kubectl get pods -n argocd -o wide
```

Expect seven pods, all `1/1 Running`, all on `minikube-m04`:

```
argocd-application-controller-0
argocd-applicationset-controller-…
argocd-dex-server-…
argocd-notifications-controller-…
argocd-redis-…
argocd-repo-server-…
argocd-server-…
```

## 3. Open the Argo CD UI

1. Get the password of the `admin` user:

   ```bash
   make argocd-password
   ```

2. Open the tunnel. The command blocks, so run it in its own terminal:

   ```bash
   make argocd-ui
   ```

3. Open `https://localhost:8443`. The browser warns about the certificate, because Argo CD signs its own. Accept the warning.
4. Log in as `admin`, with the password from step 1.

The Applications page is empty. The apps come in step 4.

## 4. Hand the apps to Argo CD

```bash
make argocd-apps
```

Expect `application.argoproj.io/root created`. Then check the apps:

```bash
make argocd-status
```

Within a minute or two, expect this:

```
NAME               SYNC STATUS   HEALTH STATUS
external-secrets   Synced        Healthy
root               OutOfSync     Progressing
vault              Synced        Progressing
```

- `vault` stays `Progressing`, because Vault starts sealed. This is correct.
- `root` waits for wave 0. The other three apps come after you configure Vault.
- Right after its first sync, `external-secrets` can show `OutOfSync` for a short time. It turns `Synced` by itself.

## 5. Configure Vault

Do steps 1, 2 and 3 of [Vault Setup](vault-setup.md):

1. Initialise and unseal Vault.
2. Write the secrets.
3. Configure Kubernetes authentication.

Then return here.

While you do this, `root` reports failed sync attempts, because `secret-store` cannot log in to Vault yet. This is expected. `root` retries without a limit, so it continues by itself when Vault is ready.

## 6. Watch the apps finish

```bash
make argocd-status
```

Waves 1 to 3 continue by themselves: `secret-store`, then `mysql`, then `student-api`. When the MySQL pod and the API pod are ready, expect this:

```
NAME               SYNC STATUS   HEALTH STATUS
external-secrets   Synced        Healthy
mysql              Synced        Healthy
root               Synced        Healthy
secret-store       Synced        Healthy
student-api        Synced        Healthy
vault              Synced        Healthy
```

The `mysql` app creates the `student-api` namespace, with its Pod Security labels.

## 7. Reach the API

Open the tunnel in its own terminal:

```bash
make k8s-port-forward
```

From another terminal:

```bash
curl http://localhost:8888/healthcheck
```

See [Postman Collection](postman.md).

## How a new build reaches the cluster

When a pull request is merged into `main`, these steps run automatically:

1. **CI job `build-test-push`** builds, tests and lints the code. Then it pushes the image `ghcr.io/…/student-rest-api:<version>`.
2. **CI job `update-image-tag`** writes `<version>` to `image.tag` in [helm/student-api/values.yaml](../helm/student-api/values.yaml). Then it commits and pushes the change.
3. **Argo CD** finds the new commit, and syncs the `student-api` app.
4. **Kubernetes** does a rolling update. The new pod is ready before the old pod stops.

See [CI/CD](ci-cd.md) for the two jobs.

**Deploy without waiting.** Argo CD looks at git at least every 3 minutes. To make it look now, click **Refresh** on the `student-api` app in the UI.

**Check which build runs:**

```bash
kubectl get deploy student-api -n student-api -o jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
```

**Pull before your next push.** After each CI run, the commit of `update-image-tag` is only on GitHub. Get it first, or git rejects your next push:

```bash
git pull --ff-only
```

## Change a setting

1. Edit the `values.yaml` of the chart, in [helm/](../helm/).
2. Commit and push.

Argo CD deploys the change in the same way. A push that changes only files under `helm/` does not run CI, because it needs no new image. A change to the `student-api` ConfigMap, for example `ginMode`, restarts the API pods.

If you change the cluster with `kubectl` or `helm upgrade`, Argo CD undoes the change (self-heal).

## Day-to-day commands

| Command | What it does |
| --- | --- |
| `make argocd-status` | Shows every app, with its sync and health status |
| `make argocd-ui` | Opens `https://localhost:8443` onto the Argo CD UI |
| `make argocd-password` | Prints the password of the `admin` user |
| `kubectl annotate application student-api -n argocd argocd.argoproj.io/refresh=normal --overwrite` | Makes Argo CD look at git now, the same as **Refresh** in the UI |
| `make k8s-port-forward` | Opens `localhost:8888` onto the API |
| `kubectl logs -n student-api deploy/student-api -f` | Shows the application logs |

Vault seals itself every time its pod restarts. After a restart, unseal it again. See step 1 of [Vault Setup](vault-setup.md).

See [Makefile Reference](makefile.md#argo-cd) for every `argocd-*` target.

## Cleanup

Remove everything, Argo CD and Vault included:

```bash
make k8s-cluster-down
```

See [Troubleshooting](troubleshooting.md) for common problems.
