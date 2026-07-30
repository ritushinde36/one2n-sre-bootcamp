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

# ---- Runtime stage ----
FROM alpine:3.20

# ca-certificates: needed for any TLS connections.
# tzdata: the app's DSN uses loc=Local, which needs the system timezone
# database to resolve correctly.
RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /app/bin/rest-api /usr/local/bin/rest-api

EXPOSE 8888

ENTRYPOINT ["/usr/local/bin/rest-api"]
