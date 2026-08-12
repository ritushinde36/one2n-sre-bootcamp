.PHONY: deps build run test test-coverage test-list test-one staticcheck hadolint newman migrate-up migrate-down migrate-status docker-build docker-migrate docker-run docker-down docker-network docker-mysql-up docker-mysql-down env-setup compose-up compose-down

IMAGE_NAME := student-rest-api
# defaults to the current git version, but can be overridden on the command
# line to build/tag a specific version instead, e.g.:
#   make docker-build VERSION=1.2.3
VERSION := $(shell git describe --tags --always --dirty)
NETWORK_NAME := student-api-net
APP_CONTAINER := student-rest-api
MYSQL_CONTAINER := student-mysql
# override on the command line to run extra instances, e.g.:
#   make docker-run APP_CONTAINER=student-rest-api02 HOST_PORT=8882
HOST_PORT := 8888

# Pulls in MYSQL_ROOT_PASSWORD/MYSQL_DATABASE from the gitignored .env.docker
# file, so no credential value ever appears in this committed Makefile.
-include .env.docker

deps:
	go mod tidy

build:
	go build -o bin/rest-api .

run: build
	./bin/rest-api

# test requires Docker to be running (spins up a real MySQL container).
test:
	go test ./...

# generates a coverage report and opens it in the browser.
# -coverpkg is needed because the tests live in the separate controllers_test
# package, not in controllers itself 
# requires Docker to be running (spins up a real MySQL container).
test-coverage:
	go test ./controllers_test/... -coverpkg=./controllers/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# lists all runnable test names (no Docker required) - use these with test-one.
test-list:
	@grep -rhoE '^func (Test[A-Za-z0-9_]+)' controllers_test/*.go | sed -E 's/^func //' | grep -v '^TestMain$$'

# runs a single test by name, e.g.: make test-one TEST=TestStudentLifecycle_PersistsAcrossRealDB
test-one:
	go test ./controllers_test/... -run $(TEST) -v

# requires staticcheck: go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck:
	staticcheck ./...

# lints the Dockerfile - requires hadolint: brew install hadolint (or see
# https://github.com/hadolint/hadolint#install)
hadolint:
	hadolint Dockerfile

# runs the Postman collection against a running server (`make run` first).
# requires newman: npm install -g newman
newman:
	newman run postman/student-api.postman_collection.json

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

# Requires .env.docker (DSN pointing at the student-mysql container, and
# LOG_FILE for where the app writes its logs on the student-logs volume) and
# the network/mysql container to already be up - see docker-network/docker-mysql-up.
docker-run: docker-network
	docker run -d --name $(APP_CONTAINER) -p $(HOST_PORT):8888 --network $(NETWORK_NAME) --env-file .env.docker -v student-logs:/var/log/app $(IMAGE_NAME):$(VERSION)

# Runs migrations to completion in a one-off container (--rm), separately
# from the app container - like a Kubernetes init container, but for plain
# Docker. Run this yourself, once, before scaling up any number of docker-run
# instances - it's deliberately NOT a docker-run dependency, so that scaling
# to N app containers doesn't mean N containers each redundantly trying to
# migrate.
docker-migrate: docker-network
	docker run --rm --network $(NETWORK_NAME) --env-file .env.docker --entrypoint migrate $(IMAGE_NAME):$(VERSION) up


# sends SIGTERM so the app's own graceful shutdown gets a chance to drain
# in-flight requests, then removes the stopped container. Logs aren't lost
# in the process - the app writes them to the student-logs volume (see
# docker-run), which outlives the container.
docker-down:
	docker stop $(APP_CONTAINER)
	docker rm $(APP_CONTAINER)

docker-network:
	docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)

docker-mysql-up: docker-network
	docker run -d --name $(MYSQL_CONTAINER) \
		--network $(NETWORK_NAME) \
		-v student-mysql-data:/var/lib/mysql \
		-p 3306:3306 \
		-e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) \
		-e MYSQL_DATABASE=$(MYSQL_DATABASE) \
		mysql:8.0

docker-mysql-down:
	docker rm -f $(MYSQL_CONTAINER)

# generates .env.docker with dev-friendly defaults, but only if it doesn't
# already exist - so a fresh clone needs zero manual file creation, and any
# file you've customized is left alone.
env-setup:
	@if [ ! -f .env.docker ]; then \
		cp .env.docker.example .env.docker; \
		echo "created .env.docker from .env.docker.example"; \
	fi

# builds the app image and starts both containers via Compose - no manual
# network/mysql/build steps needed, unlike the docker-* targets above.
# IMAGE_NAME/VERSION are exported so the image gets tagged the same way
# docker-build tags it (git version), instead of Compose's default :latest.
compose-up: env-setup
	IMAGE_NAME=$(IMAGE_NAME) VERSION=$(VERSION) docker compose --env-file .env.docker up -d --build

# stops and removes both containers (the student-mysql-data volume persists).
compose-down:
	docker compose --env-file .env.docker down
