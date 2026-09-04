# Docker & Docker Compose

← [Back to README](../README.md)

This page explains how to run the app in Docker. It covers both the manual Docker flow and Docker Compose, step by step. Use it if you do not want to install Go or MySQL directly.

## Running with Docker

The app can also be built and run as a container, without a local Go toolchain. This uses a separate MySQL container instead of a locally installed MySQL server.

**Files involved:**

- [Dockerfile](../Dockerfile): builds the app in two stages. First it compiles `rest-api` and `migrate`. Then it copies them, and the migrations folder, into a small runtime image.
- [docker-entrypoint.sh](../docker-entrypoint.sh): starts the app when the container runs. It does not run migrations. Those run separately, in step 4.
- [docker-compose.yml](../docker-compose.yml): runs the app, MySQL, and migrations together, as one Compose project. This replaces the manual steps below.
- `.env`: the same file you created in [Setup](setup.md). The Docker commands below also use it.

**1. Update `.env`.** If you have not created it yet, create it first:

```bash
cp .env.example .env
```

Set `DSN` to point at the `student-mysql` container instead of `127.0.0.1`. Also add `MYSQL_ROOT_PASSWORD` and `MYSQL_DATABASE`. The `mysql` container reads these only once, the first time it starts, to set its root password and create the `student_db` database:

```
MYSQL_ROOT_PASSWORD=your_password
MYSQL_DATABASE=student_db
MYSQL_PORT=3306
DSN=root:your_password@tcp(student-mysql:3306)/student_db?charset=utf8mb4&parseTime=True&loc=Local
PORT=8888
GIN_MODE=release
LOG_FILE=/var/log/app/rest-api.log
MIGRATE_LOG_FILE=/var/log/app/migrate.log
```

`MYSQL_PORT` controls the port on your machine, not the port inside the container:

- It is the host port that `docker-mysql-up` publishes MySQL's `3306` on.
- Change it only if something on your machine already listens on 3306 (for example, to `3307`).
- `DSN`'s port always stays `3306`. That is the container's own internal port, and `MYSQL_PORT` does not affect it.

Note: `student-mysql` is the container name the app connects to, over the Docker network created in the next step.

- That hostname resolves only for containers on that network, not from your host machine.
- If you switch back to running the app directly later, change `DSN`'s host back to `127.0.0.1`.

**2. Start a MySQL container:**

```bash
make docker-mysql-up
```

- Creates the `student-api-net` Docker network, if it does not already exist.
- Starts a `mysql:8.0` container named `student-mysql` on that network.
- Persists data in the `student-mysql-data` volume.
- Publishes port `3306` to the host, using `MYSQL_PORT` from `.env` (see above).

**3. Build the image:**

```bash
make docker-build
```

- Builds `student-rest-api:<version>`. By default, `<version>` comes from `git describe --tags --always --dirty`.
- Override it to build a specific version: `make docker-build VERSION=1.2.3`.

**4. Apply migrations:**

```bash
make docker-migrate
```

- Runs the pending migrations to completion, in a separate, temporary container, before any app container starts.
- Run this once, not once per app container.
- Its logs go to both stdout and the `student-migrate-logs` volume (`MIGRATE_LOG_FILE`), so they are not lost once the container is removed. See [Reading logs from the volume](#reading-logs-from-the-volume) below.

**5. Run the app container:**

```bash
make docker-run
```

- Starts the app container in the background, using `.env` for its settings.
- Connects to MySQL through the `DSN` from `.env`.
- Publishes port `8888`, and mounts a `student-logs` volume for its logs.
- Writes its logs to both stdout (tail with `docker logs -f student-rest-api`) and that volume, so the logs survive `make docker-down`. See [Reading logs from the volume](#reading-logs-from-the-volume) below.

**6. Verify:**

```bash
curl http://localhost:8888/healthcheck
curl http://localhost:8888/readyz
```

**Cleanup:**

```bash
make docker-down       # gracefully stops (SIGTERM) and removes the app container - logs persist in the student-logs volume
make docker-mysql-down # removes the MySQL container (the student-mysql-data volume persists)
docker network rm student-api-net
```

For the full list of Docker targets, see the [Makefile Reference](makefile.md#docker).

Note: it is fine to run the container without a `.env` file present. The app only fails if `.env` exists but cannot be read, not if it is simply missing. Docker images never include `.env` anyway, since `.dockerignore` excludes it. Environment variables passed through `--env-file` still work.

## Running with Docker Compose

[docker-compose.yml](../docker-compose.yml) replaces the manual steps above with three services: `mysql`, `migrate`, and `rest-api`.

- **Network**: Compose creates its own network, `student-api_net`. All three services share it, and can reach each other by name. There is no separate network step to run yourself.
- **Configuration**: all three services read their settings from `.env`. This is the same file the manual `docker-run`, `docker-mysql-up`, and `docker-migrate` commands use.
- **Startup order**: `migrate` runs first, once `mysql` is healthy. `rest-api` starts only after `migrate` finishes successfully.

**1. Make sure `.env` exists.** If you have not created it yet, create it first:

```bash
cp .env.example .env
```

Compose reads the same file as the manual Docker flow above. Set `DSN`'s host to `student-mysql`, and add `MYSQL_ROOT_PASSWORD` and `MYSQL_DATABASE` (see step 1 above for what these do):

```
MYSQL_ROOT_PASSWORD=yourpassword
MYSQL_DATABASE=student_db
MYSQL_PORT=3306
DSN=root:yourpassword@tcp(student-mysql:3306)/student_db?charset=utf8mb4&parseTime=True&loc=Local
PORT=8888
GIN_MODE=release
LOG_FILE=/var/log/app/rest-api.log
MIGRATE_LOG_FILE=/var/log/app/migrate.log
```

**2. Build the image, apply migrations, and start the app:**

```bash
make compose-up
```

- Before starting anything, it prints whether `mysql` is already running. This is informational only, and does not skip or change what Compose does next.
- After `up` finishes, it tails `migrate`'s own logs. This shows right away whether it applied anything, or found nothing pending, without a separate check.

Verify the same way as above (`curl http://localhost:8888/healthcheck`).

**Cleanup:**

```bash
make compose-down
```

Stops and removes all three containers. (`migrate` already exited on its own, by this point.)

Compose names its volumes by project, not with the plain names used in the manual flow:

- `student-api_student-mysql-data`
- `student-api_student-logs`
- `student-api_student-migrate-logs`

See [Reading logs from the volume](#reading-logs-from-the-volume).

For the full list of Compose targets, see the [Makefile Reference](makefile.md#docker-compose).

### Overriding Compose variables

You can override four variables. Set them in `.env`, or pass them on the command line:

```bash
make compose-up VAR=value
# or
VAR=value make compose-up
```

| Variable | Default | What it changes |
|---|---|---|
| `IMAGE_NAME` | `student-rest-api` | The name used for the image that `migrate` and `rest-api` run, for example `<name>:<version>` |
| `VERSION` | `git describe --tags --always --dirty` | The version tag for that image. Uses the same default as `make docker-build`. |
| `PORT` | `8888` | The port `rest-api` uses, both on your machine and inside the container. Change this, and the app stays reachable on the new port. |
| `MYSQL_PORT` | `3306` | The port on your machine that connects to MySQL's port 3306 inside the container. Only this changes. `DSN` still uses 3306, since that is fixed inside the container. |

```bash
# tag/run a specific version instead of the current git describe
make compose-up VERSION=1.2.3

# build/run under a different image name
make compose-up IMAGE_NAME=my-student-api

# avoid ports already taken on your machine
make compose-up PORT=9000 MYSQL_PORT=3307
```

### Reading logs from the volume

Both flows write logs to a named volume, not to the container itself. This means the logs stay even after the container is removed.

| | rest-api logs | migrate logs |
|---|---|---|
| Manual flow | `student-logs` | `student-migrate-logs` |
| Compose flow | `student-api_student-logs` | `student-api_student-migrate-logs` |

Compose adds its project name, `student-api_`, in front of each volume name.

`rest-api` writes its logs to stdout, and to `LOG_FILE` (`/var/log/app/rest-api.log`, set in `.env`).

`migrate` writes its logs to stdout, and to `MIGRATE_LOG_FILE` (`/var/log/app/migrate.log`, set in `.env`). This matters even more for `migrate`. It always runs in a temporary container that removes itself when done. Without the volume, its logs, including any failure output, would disappear the moment it exits.

### Reading a volume directly

A named volume is not a folder you can just open. To read one after its container is gone, or without stopping a running one, mount it from a disposable container:

```bash
# rest-api
docker run --rm -v student-logs:/logs alpine cat /logs/rest-api.log                     # manual flow
docker run --rm -v student-api_student-logs:/logs alpine cat /logs/rest-api.log         # compose flow

# migrate
docker run --rm -v student-migrate-logs:/logs alpine cat /logs/migrate.log              # manual flow
docker run --rm -v student-api_student-migrate-logs:/logs alpine cat /logs/migrate.log  # compose flow
```

### While the container is still running

For `rest-api`, these also work:

- `docker exec student-rest-api cat /var/log/app/rest-api.log`
- `docker compose exec rest-api cat /var/log/app/rest-api.log`
- `docker logs -f student-rest-api`
- `docker compose logs rest-api`
