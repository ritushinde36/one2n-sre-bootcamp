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
| (Kubernetes) An `ExternalSecret` reports `403 permission denied`, while the store reports `Valid` | The login works but the role has no policy. Run `vault read auth/kubernetes/role/external-secrets` — if `token_policies` is empty, the role was created with `policy=` instead of `token_policies=`. See [Secrets Management](secrets-management.md). |
| (Kubernetes) Pods stay in `CreateContainerConfigError` | The Secret they read does not exist yet, because the `ExternalSecret` has not synced. Fix the sync first; the pods start on their own afterwards. |
| (Kubernetes) `curl` to the node IP and NodePort times out | Use `make k8s-port-forward` and call `localhost:8888` instead. |
| (Kubernetes) A pod stays in `Init:Error` | A migration failed. Read `kubectl logs -n student-api deploy/rest-api -c migrate`. The API will not start until migrations succeed. |
| (Kubernetes) The API starts but `/readyz` fails | The DSN in Vault points at the wrong host. It must use `mysql.student-api.svc`, not `127.0.0.1` or a Docker network name. |
