# Secrets Management

← [Back to README](../README.md)

This page explains how the Kubernetes deployment gets its secrets. It covers Vault, the External Secrets Operator, and the steps to set both up. Do this before you deploy the app. See [Minikube Cluster](minikube.md) for the cluster itself.

## Why it works this way

The manifests in [manifests/](../manifests/) hold no secret values. The database password and the DSN live in Vault. The External Secrets Operator reads them and creates the Kubernetes Secrets that the pods use.

This keeps every secret out of Git. It also means you rotate a value in one place, and the operator syncs the change.

## How the pieces fit

```
Vault  ──read──▶  External Secrets Operator  ──creates──▶  Kubernetes Secret  ──▶  Pod
```

| Object | File | What it does |
| --- | --- | --- |
| `ClusterSecretStore` | [secret-store.yml](../manifests/secret-store.yml) | Tells the operator where Vault is and how to log in |
| `ExternalSecret` | [application.yml](../manifests/application.yml), [database.yml](../manifests/database.yml) | Names one Vault key and the Kubernetes Secret to create from it |

The store is cluster-scoped, so both `ExternalSecret` objects reference it by name alone.

## 1. Install the External Secrets Operator

The operator supplies the `ClusterSecretStore` and `ExternalSecret` resource types. Without it, `kubectl apply` fails, because those types do not exist yet.

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm repo update
helm install external-secrets external-secrets/external-secrets \
  --namespace external-secrets --create-namespace
```

The namespace and the service account name must both be `external-secrets`. [secret-store.yml](../manifests/secret-store.yml) refers to them.

Check the pods are running:

```bash
kubectl get pods -n external-secrets
```

Expect three pods, all `1/1 Running`:

```
NAME                                                READY   STATUS
external-secrets-...                                1/1     Running
external-secrets-cert-controller-...                1/1     Running
external-secrets-webhook-...                        1/1     Running
```

Wait until all three are ready before you continue:


## 2. Install Vault

```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo update
helm install vault hashicorp/vault \
  --namespace vault --create-namespace \
  -f manifests/values/vault-values.yaml
```

Check the pod:

```bash
kubectl get pods -n vault
```

Expect `vault-0` at **`0/1 Running`**. That is correct, not a failure — Vault starts sealed, and a sealed Vault reports itself as not ready. Step 3 fixes it.

[vault-values.yaml](../manifests/values/vault-values.yaml) pins Vault to the node labelled `type=dependent_services`, gives it 1Gi of storage, and fixes the data directory permissions with an init container.

Note: this runs Vault in **standalone** mode, not dev mode. Dev mode starts unsealed and loses everything on restart. Standalone mode keeps its data, but you must initialise and unseal it yourself. That is the next step.

## 3. Initialise and unseal Vault

The pod stays `0/1 Running` until Vault is unsealed. This is expected.

```bash
kubectl exec -n vault vault-0 -- vault operator init
```

**Save the output.** It prints five unseal keys and the root token. There is no way to recover them. Losing them means losing every secret in Vault.

Unseal with any three of the five keys, one command each:

```bash
kubectl exec -n vault vault-0 -- vault operator unseal <key-1>
kubectl exec -n vault vault-0 -- vault operator unseal <key-2>
kubectl exec -n vault vault-0 -- vault operator unseal <key-3>
```

The pod then reports `1/1 Running`.

Vault seals again whenever its pod restarts. You must unseal it again each time.

## 4. Write the secrets

Open a shell in the Vault pod and log in with the root token:

```bash
kubectl exec -it -n vault vault-0 -- sh
vault login <root-token>
```

Enable the key-value engine at `secret`, which is the path [secret-store.yml](../manifests/secret-store.yml) reads:

```bash
vault secrets enable -path=secret kv-v2
```

Write both values under one key, `student-api`:

```bash
vault kv put secret/student-api \
  MYSQL_ROOT_PASSWORD='<your-password>' \
  DSN='root:<your-password>@tcp(mysql.student-api.svc:3306)/student_db?charset=utf8mb4&parseTime=True&loc=Local'
```

Two points matter here:

- The DSN host must be `mysql.student-api.svc`. This is the cluster DNS name of the MySQL Service. A `127.0.0.1` or a Docker network name does not resolve inside the cluster.
- The password in the DSN and `MYSQL_ROOT_PASSWORD` must match. MySQL sets its root password from the second value, and the app connects with the first.

## 5. Configure Kubernetes authentication

The operator logs in to Vault with its service account. Vault must accept it.

Still inside the Vault pod:

```bash
vault auth enable kubernetes

vault write auth/kubernetes/config \
  kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"

vault policy write external-secrets - <<EOF
path "secret/data/student-api" {
  capabilities = ["read"]
}
EOF

vault write auth/kubernetes/role/external-secrets \
  bound_service_account_names=external-secrets \
  bound_service_account_namespaces=external-secrets \
  policy=external-secrets \
  ttl=24h
```

The role name, the service account name and the namespace must all match [secret-store.yml](../manifests/secret-store.yml). Exit the pod when you are done.

## 6. Apply the store

```bash
kubectl apply -f manifests/secret-store.yml
```

## Verify

The store must report `Valid`:

```bash
kubectl get clustersecretstore vault-backend
```

After you apply the application and database manifests, each `ExternalSecret` must report `SecretSynced`:

```bash
kubectl get externalsecret -n student-api
```

Confirm the Kubernetes Secrets now exist:

```bash
kubectl get secret -n student-api app-secret db-secret
```

## Rotating a secret

Write the new value in Vault. The operator re-reads every hour, set by `refreshInterval` in the `ExternalSecret`. To apply it at once:

```bash
kubectl annotate externalsecret app-secret -n student-api \
  force-sync=$(date +%s) --overwrite
```

The pods do not restart on their own when a Secret changes. Restart them to pick up the new value:

```bash
kubectl rollout restart deployment rest-api -n student-api
```

## Limitations

- **Vault seals on every restart.** Nobody unseals it automatically, so a node reboot leaves the cluster without secrets until someone runs the unseal commands by hand.
- **The unseal keys and root token are not stored anywhere.** Whoever runs the setup keeps them.
- **One Vault replica, one node.** There is no high availability.

See [Troubleshooting](troubleshooting.md) for what to do when a secret does not sync.
