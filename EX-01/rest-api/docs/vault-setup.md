# Vault Setup

← [Back to README](../README.md)

This page configures Vault after you install it. It covers three steps: initialise and unseal Vault, write the secrets, and let the External Secrets Operator log in.

Before you start, the Vault pod `vault-0` must exist, at `0/1 Running`.

## 1. Initialise and unseal Vault

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

Vault seals again whenever its pod restarts. After every restart, run the three unseal commands again. See [After a restart](secrets-management.md#after-a-restart--unseal-vault).

## 2. Write the secrets

Open a shell in the Vault pod:

```bash
kubectl exec -it -n vault vault-0 -- sh
```

Wait for the `/ $` prompt. The commands from here to the end of step 3 run inside the pod, not on your machine.

Log in with the root token:

```bash
vault login <root-token>
```

Expect `Success! You are now authenticated.` Without a login, every command below fails with `permission denied`.

Enable the key-value engine at `secret`. The secret store reads this path: `path` in [manifests/secret-store.yml](../manifests/secret-store.yml), and `vault.path` in [helm/secret-store/values.yaml](../helm/secret-store/values.yaml).

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

## 3. Configure Kubernetes authentication

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

The role name, the service account name and the namespace must all match the secret store: `role` and `serviceAccountRef` in [manifests/secret-store.yml](../manifests/secret-store.yml), or `vault.role` and `serviceAccount` in [helm/secret-store/values.yaml](../helm/secret-store/values.yaml). A mismatch in any of the three is the usual reason a secret never syncs.

Exit the pod when you are done.

## Next

Return to the page you came from:

- Manifests: [Secrets Management, step 4](secrets-management.md#4-apply-the-store).
- Helm charts: [Helm Charts, step 4](helm.md#4-deploy-the-application).
- Argo CD: [Argo CD, step 6](argocd.md#6-watch-the-apps-finish).
