# Minikube Cluster

← [Back to README](../README.md)

This page explains how to set up the local Kubernetes cluster this app deploys to. It covers creating a multi-node cluster and labeling each node by role.

## Create the Cluster

Create a 4-node cluster: 1 control-plane node and 3 worker nodes.

```bash
minikube start --nodes 4 --driver=docker
```

Wait for all nodes to be in a ready state:

```bash
kubectl wait --for=condition=Ready nodes --all --timeout=120s
```

Confirm the node names:

```bash
kubectl get nodes -o wide
```

## Label the Worker Nodes

Each worker node gets a `type` label. Workloads use this label to target a specific node.

```bash
kubectl label node minikube-m02 type=application
kubectl label node minikube-m03 type=database
kubectl label node minikube-m04 type=dependent_services
```

| Node | Label | Role |
|---|---|---|
| `minikube-m02` | `type=application` | Runs the REST API |
| `minikube-m03` | `type=database` | Runs MySQL |
| `minikube-m04` | `type=dependent_services` | Runs supporting services |

Verify the labels:

```bash
kubectl get nodes -L type
```
