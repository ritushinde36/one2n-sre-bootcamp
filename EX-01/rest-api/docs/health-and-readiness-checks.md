# Health & Readiness Checks

← [Back to README](../README.md)

This page explains the app's two health probes. It covers what each one checks, and what a failure means. Read this to understand how the app reports its own health.

The app exposes two distinct probes. These match the liveness and readiness split used by container orchestrators, like Kubernetes:

- **`GET /healthcheck`**: liveness. Confirms only that the process is alive, and can respond to HTTP. It checks no external dependency. A failure here means the process itself is broken, and needs a restart.
- **`GET /readyz`**: readiness. Confirms the app can currently serve real traffic, including a database ping (with a 2-second timeout). A failure here means traffic should stop routing here. The process itself does not need a restart. Restarting does not fix a database that is down.
