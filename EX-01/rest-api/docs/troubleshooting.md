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
