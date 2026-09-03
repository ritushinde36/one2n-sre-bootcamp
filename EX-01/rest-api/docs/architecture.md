# Architecture

← [Back to README](../README.md)

This page explains how the main pieces of the system work together. It covers the client, the REST API, the database, and migrations.

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
