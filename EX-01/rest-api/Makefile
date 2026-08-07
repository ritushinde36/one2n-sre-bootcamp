.PHONY: deps build run test test-coverage test-list test-one staticcheck newman migrate-up migrate-down migrate-status docker-build docker-run docker-down docker-network docker-mysql-up docker-mysql-down

IMAGE_NAME := student-rest-api
VERSION := $(shell git describe --tags --always --dirty)
NETWORK_NAME := student-api-net
APP_CONTAINER := student-rest-api
MYSQL_CONTAINER := student-mysql

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

# Requires .env.docker (DSN pointing at the student-mysql container) and the
# network/mysql container to already be up - see docker-network/docker-mysql-up.
docker-run: docker-network
	docker run -d --name $(APP_CONTAINER) -p 8888:8888 --network $(NETWORK_NAME) --env-file .env.docker $(IMAGE_NAME):$(VERSION)

# sends SIGTERM so the app's own graceful shutdown gets a chance to drain
# in-flight requests, saves its logs (removing the container would discard
# them otherwise), then removes the stopped container.
docker-down:
	docker stop $(APP_CONTAINER)
	mkdir -p logs
	docker logs $(APP_CONTAINER) > logs/$(APP_CONTAINER)-$$(date +%Y%m%d-%H%M%S).log 2>&1
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
