# Architecture

← [Back to README](../README.md)

This page explains how the main pieces of the system work together. It covers the client, the REST API, the database, and migrations.

This is the general architecture. It holds for every way you run the app. Each deployment adds its own pieces around it — more copies of the API, a load balancer, a cluster — but the relationships below do not change.

```
Client  --  HTTP request / JSON response  -->  REST API  --  SQL query / rows  ---->  MySQL
                                                                                        ^
                                                                                        |
                                                                              schema changes
                                                                                        |
                                                                                 Migrations
```

**Client → REST API.** A client sends an HTTP request with a JSON body to an endpoint under `/api/v1` (see [API Reference](api-reference.md)). The REST API validates the request, then responds with a JSON body and an HTTP status code.

**REST API → Database.** The REST API uses GORM to turn each validated request into a query against MySQL. It returns the result as JSON. See [Data Model](data-model.md) for the schema it reads and writes.

**Migrations → Database.** A separate migration tool applies schema changes to the database. The REST API does not apply them. Migrations run separately, before the REST API starts. See [Database Migrations](migrations.md) for details.

## Architecture per deployment

Each page below shows how these pieces are arranged for that deployment, and what it adds.

| Where you run it | What the architecture adds | Page |
| --- | --- | --- |
| Your machine | Nothing. One process, one local database. | [Local Setup](setup.md) |
| Docker Compose | Each piece in its own container, on one network. | [Docker & Docker Compose](docker.md) |
| A Vagrant VM | Two API containers behind an nginx load balancer, inside one VM. | [Vagrant VM](vagrant.md) |
| Kubernetes | Pods on labelled nodes, with secrets from Vault. | [Minikube Cluster](minikube.md) |

Two things change with the deployment, and are worth reading about where they differ:

- **How migrations run.** A separate container in Docker and Compose, an init container in Kubernetes. See [Database Migrations](migrations.md).
- **Where secrets come from.** A `.env` file everywhere except Kubernetes, where Vault holds them. See [Secrets Management](secrets-management.md).
