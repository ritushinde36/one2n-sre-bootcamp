.PHONY: build run test migrate-up migrate-down migrate-status docker-build docker-run docker-network docker-mysql-up docker-mysql-down

IMAGE_NAME := student-rest-api
VERSION := $(shell git describe --tags --always --dirty)
NETWORK_NAME := student-api-net
MYSQL_CONTAINER := student-mysql

# Pulls in MYSQL_ROOT_PASSWORD/MYSQL_DATABASE from the gitignored .env.docker
# file, so no credential value ever appears in this committed Makefile.
-include .env.docker

build:
	go build -o bin/rest-api .

run: build
	./bin/rest-api

# test requires Docker to be running (spins up a real MySQL container).
test:
	go test ./...

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
	docker run -p 8888:8888 --network $(NETWORK_NAME) --env-file .env.docker $(IMAGE_NAME):$(VERSION)

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
