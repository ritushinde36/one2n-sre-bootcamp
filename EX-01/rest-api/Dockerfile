# ---- Build stage ----
FROM golang:1.26-alpine AS build

WORKDIR /app

# Copy only dependency manifests first so `go mod download` is cached as its
# own layer, independent of source code changes.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source.
COPY . .

# CGO_ENABLED=0 produces a fully static binary with no C runtime dependency,
# which is what lets the runtime stage below use a minimal base image.
# -ldflags="-s -w" strips debug symbols to shrink the binary itself.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/rest-api .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/migrate ./cmd/migrate

# ---- Runtime stage ----
FROM alpine:3.20

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=build /app/bin/rest-api /usr/local/bin/rest-api
COPY --from=build /app/bin/migrate /usr/local/bin/migrate
# cmd/migrate reads migration files off disk at runtime (not embedded), relative to the CWD.
COPY --from=build --chown=app:app /app/migrations ./migrations
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

USER app

EXPOSE 8888

# Liveness only - checks the process itself responds, not the database (see
# /readyz for that). Keeping DB reachability out of this check matters:
# HEALTHCHECK failures are what an orchestrator restarts the container over,
# and restarting rest-api can't fix a downed database.
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://localhost:8888/healthcheck || exit 1

# Starts the API server - migrations are a separate one-off step run before
# this container starts (see `make docker-migrate`), not done here on every
# boot. To run migrate directly (e.g. `up`/`status`/`down`), override the
# entrypoint:
#   docker run --rm --entrypoint migrate --network ... --env-file .env.docker <image> status
ENTRYPOINT ["docker-entrypoint.sh"]
