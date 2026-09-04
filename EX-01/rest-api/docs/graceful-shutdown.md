# Graceful Shutdown

← [Back to README](../README.md)

This page explains how the app shuts down. It covers what happens on `SIGINT` or `SIGTERM`, and why it matters. Read this to understand how the app avoids dropping requests.

On `SIGINT` or `SIGTERM`, the app stops accepting new connections. It then gives in-flight requests up to 10 seconds to finish. Then it exits. See [main.go](../main.go).

This matters for zero-downtime deploys and container restarts. Without it, those would kill the process outright, and drop any request still in progress.
