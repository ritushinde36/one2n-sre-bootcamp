# Secrets Management

← [Back to README](../README.md)

This page explains how the Kubernetes deployment gets its secrets. It covers Vault, the External Secrets Operator, and the steps to set both up. Do this before you deploy the app. See [Minikube Cluster](minikube.md) for the cluster itself.

With the Helm charts, follow [Helm Charts](helm.md) instead. That page has every step.

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



## 1. Install the External Secrets Operator

The operator supplies the `ClusterSecretStore` and `ExternalSecret` resource types. Without it, `kubectl apply` fails, because those types do not exist yet.

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm repo update
helm install external-secrets external-secrets/external-secrets \
  --namespace external-secrets --create-namespace \
  -f manifests/values/external-secrets-values.yaml \
  --version 2.11.0
```

The namespace and the service account name must both be `external-secrets`. [secret-store.yml](../manifests/secret-store.yml) refers to them.

[external-secrets-values.yaml](../manifests/values/external-secrets-values.yaml) pins all three operator pods to the node labelled `type=dependent_services`. The chart has a separate `nodeSelector` for each pod, so the file sets all three.

Check the pods are running:

```bash
kubectl get pods -n external-secrets -o wide
```

Expect three pods, all `1/1 Running` on `minikube-m04`:

```
NAME                                                READY   STATUS    NODE
external-secrets-...                                1/1     Running   minikube-m04
external-secrets-cert-controller-...                1/1     Running   minikube-m04
external-secrets-webhook-...                        1/1     Running   minikube-m04
```

Wait until all three are ready before you continue:

```bash
kubectl wait --for=condition=Ready pods --all -n external-secrets --timeout=120s
```


## 2. Install Vault

```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo update
helm install vault hashicorp/vault \
  --namespace vault --create-namespace \
  -f manifests/values/vault-values.yaml \
  --version 0.34.1
```

Check the pod:

```bash
kubectl get pods -n vault
```

Expect `vault-0` at **`0/1 Running`**. That is correct, not a failure — Vault starts sealed, and a sealed Vault reports itself as not ready. Step 3 fixes it.

[vault-values.yaml](../manifests/values/vault-values.yaml) pins Vault to the node labelled `type=dependent_services`, gives it 1Gi of storage, and fixes the data directory permissions with an init container. It also turns off the Vault agent injector. This setup does not use the injector, because the External Secrets Operator delivers the secrets.

Note: this runs Vault in **standalone** mode, not dev mode. Dev mode starts unsealed and loses everything on restart.

## 3. Configure Vault

Vault starts sealed and empty. Follow [Vault Setup](vault-setup.md). It has three steps:

1. Initialise and unseal Vault.
2. Write the secrets.
3. Configure Kubernetes authentication.

Then return here for step 4.

## 4. Apply the store

```bash
kubectl apply -f manifests/secret-store.yml
```

## 5. Verify

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


If you no longer have three unseal keys, the data in Vault cannot be recovered. You have to delete Vault's storage and set it up again — see [Limitations](#limitations).

## Limitations

- **Vault seals on every restart.** Nobody unseals it automatically, so a node reboot leaves the cluster without secrets until someone runs the unseal commands by hand.
- **The unseal keys and root token are not stored anywhere.** Whoever runs the setup keeps them.
- **One Vault replica, one node.** There is no high availability.

See [Troubleshooting](troubleshooting.md) for what to do when a secret does not sync.
