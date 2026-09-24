# Secrets Management

← [Back to README](../README.md)

This page explains how the Kubernetes deployment gets its secrets. It covers Vault, the External Secrets Operator, and the steps to set both up. Do this before you deploy the app. See [Minikube Cluster](minikube.md) for the cluster itself.

## Why it works this way

The manifests in [manifests/](../manifests/) hold no secret values. The database password and the DSN live in Vault. The External Secrets Operator reads them and creates the Kubernetes Secrets that the pods use.


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

Note: this runs Vault in **standalone** mode, not dev mode. Dev mode starts unsealed and loses everything on restart.

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

Vault seals again whenever its pod restarts, and you must unseal it again each time. See [After a restart](#after-a-restart--unseal-vault).

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

The operator logs in to Vault with its Kubernetes service account. Vault does not accept that by default. These four commands teach it to.

Run them inside the Vault pod, one at a time.

**1. Turn on the Kubernetes login method.**

Vault supports several ways to log in. This one accepts a Kubernetes service account token as proof of identity. It is off until you enable it.

```bash
vault auth enable kubernetes
```

**2. Tell Vault where the Kubernetes API is.**

When the operator presents a token, Vault asks the Kubernetes API whether that token is genuine. Vault needs the address to ask.

```bash
vault write auth/kubernetes/config \
  kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"
```

`KUBERNETES_PORT_443_TCP_ADDR` is a variable that Kubernetes sets inside every pod. It holds the API server's address. This is why the command runs inside the pod, and not from your machine.

**3. Create a policy that grants read on one path.**

A policy is a list of what a caller may do. This one allows reading the `student-api` secret, and nothing else. The operator cannot read any other secret, and cannot write at all.

```bash
vault policy write external-secrets - <<EOF
path "secret/data/student-api" {
  capabilities = ["read"]
}
EOF
```

Note the `data/` in the path. A key-value version 2 engine stores each secret under `data/`, so `secret/student-api` is written as `secret/data/student-api` in a policy.

**4. Bind the policy to one service account.**

A role connects the two: a named service account, in a named namespace, receives a named policy. Any other account that tries to log in is refused.

```bash
vault write auth/kubernetes/role/external-secrets \
  bound_service_account_names=external-secrets \
  bound_service_account_namespaces=external-secrets \
  token_policies=external-secrets \
  token_ttl=24h
```

Check the role took the policy:

```bash
vault read auth/kubernetes/role/external-secrets
```

`token_policies` must list `external-secrets`. An empty `[]` means the parameter name was wrong.

The role name, the service account name and the namespace must all match [secret-store.yml](../manifests/secret-store.yml). A mismatch in any of the three is the usual reason a secret never syncs.

Exit the pod when you are done.

## 6. Apply the store

```bash
kubectl apply -f manifests/secret-store.yml
```

## 7. Verify

One check at this point. The store must report `Valid`:

```bash
kubectl get clustersecretstore vault-backend
```

```
NAME            STATUS   CAPABILITIES   READY
vault-backend   Valid    ReadWrite      True
```

`Valid` means the operator reached Vault and logged in. 


Secrets are now set up. Go back to [Minikube Cluster](minikube.md) and deploy the application.

### After you deploy

These two checks only work once the app is running, because the application and database manifests create the namespace and the `ExternalSecret` objects. Run them too early and you get "No resources found" and "namespaces not found" — which is expected, not a failure.

Each `ExternalSecret` must report `SecretSynced`:

```bash
kubectl get externalsecret -n student-api
```

And the Kubernetes Secrets it creates must exist:

```bash
kubectl get secret -n student-api app-secret db-secret
```

---

The setup is complete. The sections below are reference, not steps.

## After a restart — unseal Vault

**Read this first when you come back to a cluster you set up earlier.**

Vault seals itself every time its pod stops. A machine reboot, `minikube stop`, a node restart — all of them. Nothing unseals it for you, and you do this every time.

You will see:

- `vault-0` back at `0/1 Running`
- every `ExternalSecret` at `SecretSyncedError`
- **the app still serving normally**

That last one is why this is easy to miss. The Kubernetes Secrets already created do not disappear when Vault seals, so the running app carries on. Nothing looks broken until a secret needs to change, or a pod restarts and cannot find its credentials.

Unseal with any three of your five keys:

```bash
kubectl exec -n vault vault-0 -c vault -- vault operator unseal <key-1>
kubectl exec -n vault vault-0 -c vault -- vault operator unseal <key-2>
kubectl exec -n vault vault-0 -c vault -- vault operator unseal <key-3>
```

Confirm it opened:

```bash
kubectl exec -n vault vault-0 -c vault -- vault status
```

`Sealed` reads `false`, and the pod returns to `1/1 Running`.

Then force the secrets to re-sync, rather than waiting up to an hour for the next refresh:

```bash
kubectl annotate externalsecret app-secret -n student-api force-sync=$(date +%s) --overwrite
kubectl annotate externalsecret db-secret  -n student-api force-sync=$(date +%s) --overwrite
kubectl get externalsecret -n student-api
```

Both must return to `SecretSynced`.

If you no longer have three unseal keys, the data in Vault cannot be recovered. You have to delete Vault's storage and set it up again — see [Limitations](#limitations).

## Limitations

- **Vault seals on every restart.** Nobody unseals it automatically, so a node reboot leaves the cluster without secrets until someone runs the unseal commands by hand.
- **The unseal keys and root token are not stored anywhere.** Whoever runs the setup keeps them.
- **One Vault replica, one node.** There is no high availability.

See [Troubleshooting](troubleshooting.md) for what to do when a secret does not sync.
