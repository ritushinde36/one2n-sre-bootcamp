# Troubleshooting

← [Back to README](../README.md)

This page lists common errors and how to fix them. Use it when something does not work as expected.

| Issue | Solution |
|---|---|
| App exits immediately with "environment variable DSN is not set" | Create `.env` (`cp .env.example .env`), and fill in a real DSN. |
| App exits with "failed to connect to database" | Check that MySQL is running, the DSN's host, port, and credentials are correct, and the database exists. See [Setup](setup.md) step 1. |
| Docker container fails to connect to MySQL, or connects to the wrong database | Set `DSN`'s host correctly in `.env`: `student-mysql` for Docker, `127.0.0.1` for running directly. See [Docker & Docker Compose](docker.md) step 1. |
| `make test` hangs or fails to start | Start Docker. The integration suite needs it, to launch its MySQL Testcontainer. |
| `/readyz` returns 503 | Check that MySQL is running, and reachable from wherever the app is deployed. |
| Migration `up` fails with "refusing to proceed: pre-existing students table does not match migration 00001" | Compare the printed `existing` and `expected` DDL, and reconcile them by hand. The migration tool will not alter a mismatched table for you. |
| Creating or updating a student returns 400, mentioning an unexpected field name (for example `id`, or `created_at`) | Remove that field. The API accepts only `name`, `email`, `age`, `class`, and `department`. |
| (Vagrant VM) Port 8080 already in use on the host | Free port 8080, then run `make vagrant-up` again. |
| (Vagrant VM) `.env` not found errors from Compose | Confirm `.env` exists before `make vagrant-up`, and that provisioning ran. |
| (Vagrant VM) The VM does not finish its first boot | UTM shows permission pop-ups during boot. Approve them. |
| (Vagrant VM) You need the provisioning output after the run finished | Read the newest `logs/provision-vm-*.log` file. These files are on the host, and they survive `make vagrant-destroy`. |
| (Kubernetes) A pod stays `Pending` | No node carries the label it needs. Run `kubectl get nodes -L type` and label the nodes. See [Minikube Cluster](minikube.md). |
| (Kubernetes) `kubectl apply` fails with "no matches for kind ExternalSecret" | The External Secrets Operator is not installed. See [Secrets Management](secrets-management.md). |
| (Kubernetes) The Vault pod stays `0/1 Running` | Vault is sealed. Unseal it with three keys. It seals again after every restart. See [Secrets Management](secrets-management.md). |
| (Kubernetes) An `ExternalSecret` never reaches `SecretSynced` | Run `kubectl describe externalsecret -n student-api`. Usually Vault's Kubernetes auth role does not match the name or namespace in `secret-store.yml`. |
| (Kubernetes) An `ExternalSecret` was synced but now shows `SecretSyncedError` | Vault is sealed, so the refresh failed. The pods keep running on the Secret that already exists, but no new value can sync. Unseal Vault. |
| (Kubernetes) An `ExternalSecret` reports `403 permission denied`, while the store reports `Valid` | The login works but the role has no policy. Run `vault read auth/kubernetes/role/external-secrets` — if `token_policies` is empty, the role was created with `policy=` instead of `token_policies=`. See [Vault Setup](vault-setup.md#3-configure-kubernetes-authentication). |
| (Kubernetes) Pods stay in `CreateContainerConfigError` | The Secret they read does not exist yet, because the `ExternalSecret` has not synced. Fix the sync first; the pods start on their own afterwards. |
| (Kubernetes) `minikube service student-api` reports that the service has no node port | The service is `ClusterIP` on purpose. Use `make k8s-port-forward` and call `localhost:8888` instead. |
| (Kubernetes) `curl localhost:8888` fails with "connection refused" | The tunnel is not running. Start `make k8s-port-forward` in its own terminal, and leave it open. |
| (Kubernetes) A pod stays in `Init:Error` | A migration failed. Read `kubectl logs -n student-api deploy/rest-api -c migrate`. With the Helm charts, the deployment is `deploy/student-api`. The API will not start until migrations succeed. |
| (Kubernetes) The API starts but `/readyz` fails | The DSN in Vault points at the wrong host. It must use `mysql.student-api.svc`, not `127.0.0.1` or a Docker network name. |
| (Kubernetes) Old students, or an old MySQL password, come back after `make k8s-down` and `make k8s-deploy` | `make k8s-down` does not delete the MySQL data. See [Cleanup](minikube.md#cleanup). |
| (Helm) `make helm-deploy` fails with "no matches for kind ClusterSecretStore" or "ExternalSecret" | The External Secrets Operator is not installed. Run `make helm-dependent-services-up` first. |
| (Helm) `make helm-deploy` times out while it waits for `mysql` or `student-api` | The pods did not become ready. Usually a secret did not sync, because Vault is sealed or not configured. Run `make helm-status` and `kubectl describe externalsecret -n student-api`. Fix the cause, then run `make helm-deploy` again. |
| (Helm) A setting under `vault:` or `external-secrets:` has no effect | The key must be the exact dependency name in `Chart.yaml`. Helm ignores a misspelled key and gives no error. Check with `helm template vault helm/vault -n vault`. |
| (Helm) `helm upgrade` of `mysql` fails with "Forbidden: updates to statefulset spec" | You changed `storage.size`, or another part of the volume claim template. Kubernetes refuses that change after the first install. Set the old value again. |
| (Helm) A value from `--set` is gone after `make helm-deploy` | This is on purpose. The Make targets use `--reset-values`, so each run deploys the values in git. To keep a value, put it in the `values.yaml` of the chart. See [Change a setting](helm.md#change-a-setting). |
| (Argo CD) `root` stays at "waiting for healthy state of argoproj.io/Application/vault" | Vault is sealed. Do steps 1 to 3 of [Vault Setup](vault-setup.md). `root` continues by itself. |
| (Argo CD) `root` reports "one or more synchronization tasks completed unsuccessfully. Retrying attempt …" | `secret-store` cannot log in to Vault yet. Finish [Vault Setup](vault-setup.md). `root` retries without a limit. |
| (Argo CD) `secret-store` stays `Degraded` | The store cannot log in to Vault. Check that Vault is unsealed, and that the role, service account and namespace match `helm/secret-store/values.yaml`. See step 3 of [Vault Setup](vault-setup.md). |
| (Argo CD) A commit is on GitHub, but the app still shows the old revision | Argo CD looks at git at least every 3 minutes. Click **Refresh** on the app in the UI. Also check that the `targetRevision` of the app is the branch you pushed to. |
| (Argo CD) A change made with `kubectl` or `helm upgrade` disappears | This is on purpose. Self-heal returns the cluster to git. Make the change in git instead. See [Change a setting](argocd.md#change-a-setting). |
| (CI) `update-image-tag` fails at "Commit and push" with a permission error | The Actions secret `DEPLOY_KEY` is missing, or the deploy key has no write access. See [CI/CD](ci-cd.md). |
| (CI) `update-image-tag` fails on `main` with "GH006: Protected branch update failed" | The ruleset on `main` does not let deploy keys bypass the pull request rule. See [CI/CD](ci-cd.md). |
